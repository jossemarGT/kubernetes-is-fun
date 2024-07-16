import asyncio
import kopf
import kubernetes
import requests
import base64
import urllib3


@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT},
               annotations={'secret-handshake-done': kopf.ABSENT})
async def handle_legacy_pod(event, status, namespace, name, logger, **_):
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    podIP = status.get('podIP', '')
    if not podIP:
        return

    logger.info(f"=== Found {name} in {namespace} namespace w/ {podIP} address")

    asyncio.create_task(secret_handshake(podIP, namespace, name, logger))


async def secret_handshake(podIP, namespace, name, logger):
    try:
        logger.info(f"=== Pod secret handshake starts")

        r = requests.get(f"http://{podIP}:3000/internal/key", timeout=30)
        if r.status_code != 200:
            logger.error("handshake failed fetching the key")
            return

        key = base64.b64decode(r.text)
        secret = base64.b64encode(key + "-acknowledge".encode())

        r = requests.post(
            f"http://{podIP}:3000/internal/secret", data=secret, timeout=30)

        if r.status_code < 200 or r.status_code > 299:
            logger.error(f"handshake failed fetching the key. Response: {r.text}")
            return

        kubernetes.config.load_incluster_config()
        api = kubernetes.client.CoreV1Api()
        api.patch_namespaced_pod(name=name,namespace=namespace, body=[{
            'op': 'add',
            'path': '/metadata/annotations',
            'value': {
                'secret-handshake-done': 'true'
            }
        }])

        logger.info(f"=== Pod secret handshake finished")

    except requests.exceptions.Timeout:
        logger.info(f"=== Connection timeout http://{podIP}:3000/")
    except urllib3.exceptions.NewConnectionError:
        logger.info("Unable to stablish new connection. Will retry")
    except asyncio.CancelledError:
        logger.info(f"=== Pod secret handshake is cancelled!")
    except Exception as e:
        logger.error(e)
