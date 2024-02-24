## Paso a paso

> ℹ️ **Recuerda**
> La solución final está en `controller.py`, y por ello todo lo que leas acá lo
> haremos en otro archivo llamado `ctrl.py`

Si estás siguiendo el laboratorio en vivo puedes utilizar este "paso a paso". 

0. Levanta un ambiente virtual de python e instala todas las librerías que
   usaremos en nuestras pruebas

```python
python -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements.txt
```

1. Creamos nuestro controlador que escuchará todos los eventos de tipo `pod`

```python
import kopf

@kopf.on.event('pods')
async def pod_in_sight(status, namespace, name, logger, **_):
    logger.info( f"=== Found {name} in {namespace} namespace")
```

2. Corremos nuestra lógica y veamos que sucede (esto lo debes hacer cada vez que modifiques `ctrl.py`)

```shell
kopf run --all-namespaces ctrl.py
```

3. Vamos a reducir nuestro alcance a solo los Pod que estén corriendo

```python
import kopf

@kopf.on.event('pods')
async def pod_in_sight(status, namespace, name, logger, **_):
    # https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    logger.info( f"=== Found {name} in {namespace} namespace w/status {status}")
```

4. Delimitamos aún más con solo los Pod que nos interesan por medio de Label(s)

```python
import kopf

@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT},)
async def pod_in_sight(status, namespace, name, logger, **_):
    # https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    logger.info( f"=== Found {name} in {namespace} namespace w/status {status}")
```

5. Ahora tomamos lo más importante, listar la IP que se le asigna al Pod

```python
import kopf

@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT},)
async def pod_in_sight(status, namespace, name, logger, **_):
    # https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    podIP = status.get('podIP', '')
    if not podIP:
        return

    logger.info(
        f"=== Found {name} in {namespace} namespace w/ {podIP} address")
```

6. Agregamos la lógica de "secret handshake"

```python
import asyncio
import kopf
import kubernetes
import requests
import base64
import urllib3

@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT})
async def pod_in_sight(event, status, namespace, name, logger, **_):
    # https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    podIP = status.get('podIP', '')
    if not podIP:
        return

    logger.info(
        f"=== Found {name} in {namespace} namespace w/ {podIP} address")

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

        logger.info(f"=== Pod secret handshake finished")

    except requests.exceptions.Timeout:
        logger.info(f"=== Connection timeout http://{podIP}:3000/")
    except urllib3.exceptions.NewConnectionError:
        logger.info("Unable to stablish new connection. Will retry")
    except asyncio.CancelledError:
        logger.info(f"=== Pod secret handshake is cancelled!")
    except Exception as e:
        logger.error(e)
```

7. Add annotation as verification

```python
import asyncio
import kopf
import kubernetes
import requests
import base64
import urllib3


@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT},
               annotations={'secret-handshake-done': kopf.ABSENT})
async def pod_in_sight(event, status, namespace, name, logger, **_):
    # https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    podIP = status.get('podIP', '')
    if not podIP:
        return

    logger.info(
        f"=== Found {name} in {namespace} namespace w/ {podIP} address")

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
```

8. En este punto ya el código es igual a la solución final. ¿te dite cuenta que
   estamos escribiendo un `Annotation`?
