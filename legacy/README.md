# Legacy Automation (Operator)

A veces encontramos que manejar ciertos procesos o aplicaciones requiere
demasiada intervención manual. Veamos cómo podemos extender Kubernetes para
automatizar la solución a estos problemas.

## El concepto

Se le dice
[operador](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) a la
pieza de software que permite ampliar o automatizar el comportamiento de un
clúster sin modificar el código de Kubernetes al vincular los controladores a
uno o más recursos personalizados.

Es decir un operador es el software/aplicación creada con el objetivo de
ejecutar las tareas que su contraparte humana (human operator) puede realizar en
un clúster de Kubernetes.

## El ejercicio a resolver

Tu equipo recién terminó la ardua tarea de migrar todas las aplicaciones a
[contedores](https://glossary.cncf.io/es/container/), sin embargo hay una que
necesita un trato especial para ser desplegada. Muchos la llaman "Legacy" porque
nadie conoce al autor ni el lenguaje en que está escrita pero existe desde mucho
tiempo atrás.

Legacy tiene la peculiaridad de requerir un "apretón de mano secreto" antes de
iniciar a operar, el cual no es sencillo de automatizar con lo que Kubernetes
ofrece por defecto. Este proceso se puede resumir en los siguientes pasos:

- Obtenemos una mitad de la cadena secret accediendo al endpoint `/internal/key`.
- Unir esta cadena con la otra mitad que solo el equipo de operaciones conoce.
- Enviamos la nueva cadena secreta al endpoint `/internal/secret`.
- Si hicimos el proceso bien, Legacy comenzará a funcionar.

### El despliegue manual

Legacy ya puede ser desplegada como
[Deployment](https://kubernetes.io/es/docs/concepts/workloads/controllers/deployment/),
con su respectivos
[Service](https://kubernetes.io/es/docs/concepts/services-networking/service/) e
[Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/). Aún
así requiere que se inicialize manualmente como lo describimos anteriormente; en
código el proceso se vería de la siguiente manera:

- Construimos la imagen de Legacy

```shell
../_hack/build.sh ./mock-app
```

- Aplicamos su manifest en Kubernetes

```shell
kubectl apply -f manifests/legacy-mock.yaml
```

- Verificamos si Legacy está corriendo

```shell
$  kubectl get deployments
NAME          READY   UP-TO-DATE   AVAILABLE   AGE
legacy-mock   1/1     1            0           15s
```

- Verificamos sus endpoints (recuerda esto necesita `minikube tunnel`)

```shell
$ curl -i http://localhost/health
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 16
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

{"status":"ok"}
$ curl -i http://localhost/
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 16
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

Uninitialized
```

- Tomamos la mitad de la cadena secreta

```shell
$ curl http://localhost/internal/key
VGhpcyBpcyBAICQzY3IzdCBrZXk=
```

- Generamos la cadena completa de nuestro lado

```shell
$ curl -s http://localhost/internal/key | base64 -d | xargs -I {} echo '{}-acknowledge' | base64
VGhpcyBpcyBAICQzY3IzdCBrZXktYWNrbm93bGVkZ2UK
```

- Enviamos la cadena completa de vuelta a Legacy

```shell
$ curl -i -XPOST -d 'VGhpcyBpcyBAICQzY3IzdCBrZXktYWNrbm93bGVkZ2UK' http://localhost/internal/secret
HTTP/1.1 202 Accepted
Content-Type: text/plain; charset=utf-8
Content-Length: 9
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

Accepted
```

- Por ultimo revisamos si Legacy ya está sirviendo los endpoints públicos

```shell
$ curl -i http://localhost/
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 16
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

It works!
```

Definitivamente no es algo que querramos hacer manualmente en cada despliegue.

## La solución

Utilizaremos el
[Kubernetes Operator Pythonic Framework](https://github.com/nolar/kopf) ó _KOPF_
para escribir la lógica del proceso que necesitamos automatizar en un
lenguaje de programación sencillo de comprender.

0. En un directorio nuevo, creamos un ambiente virtual de python

```shell
mkdir operator && cd operator
python -m venv .venv
source .venv/bin/activate
```

1. Instalamos `kopf` como cli y librería junto a sus dependencias

```shell
pip install kopf
cat << EOF >requirements.txt
kopf
kubernetes
pyyaml
requests
EOF
python -m pip install -r requirements.txt
```

3. Creamos un archivo llamado `legacy-automation.py` donde irá la lógica de
   nuestro operador

4. Registramos la función que debe ejecutarse por cada evento de `Pod` que se
   encuentra en cluster.

```python
import kopf

@kopf.on.event('pods')
async def handle_legacy_pod(event, status, namespace, name, logger, **_):
  logger.info(f"Found {name} in {namespace} namespace")
```

5. Probamos la lógica que llevamos de nuestro operador, ejecutandolo en "modo
   dev" con el cli de `kopf` (Para detenerlo usa Ctrl + C)

```shell
$ kopf run --all-namespaces legacy-automation.py
[... logs de inicialización ...]
[INFO    ] Initial authentication has finished.
[INFO    ] [kube-system/coredns-7db6d8ff4d-drgwx] Found coredns-7db6d8ff4d-drgwx in kube-system namespace
[INFO    ] [kube-system/coredns-7db6d8ff4d-drgwx] Handler 'handle_legacy_pod' succeeded.
[INFO    ] [kube-system/etcd-k8s-fun] Found etcd-k8s-fun in kube-system namespace
[INFO    ] [kube-system/etcd-k8s-fun] Handler 'handle_legacy_pod' succeeded.
[INFO    ] [kube-system/kube-apiserver-k8s-fun] Found kube-apiserver-k8s-fun in kube-system namespace
[INFO    ] [kube-system/kube-apiserver-k8s-fun] Handler 'handle_legacy_pod' succeeded.
[INFO    ] [kube-system/kube-controller-manager-k8s-fun] Found kube-controller-manager-k8s-fun in kube-system namespace
[INFO    ] [kube-system/kube-controller-manager-k8s-fun] Handler 'handle_legacy_pod' succeeded.
[INFO    ] [kube-system/kube-proxy-v2dsr] Found kube-proxy-v2dsr in kube-system namespace
[INFO    ] [kube-system/kube-proxy-v2dsr] Handler 'handle_legacy_pod' succeeded.
[INFO    ] [kube-system/kube-scheduler-k8s-fun] Found kube-scheduler-k8s-fun in kube-system namespace
[INFO    ] [kube-system/kube-scheduler-k8s-fun] Handler 'handle_legacy_pod' succeeded.
[INFO    ] [kube-system/storage-provisioner] Found storage-provisioner in kube-system namespace
[INFO    ] [kube-system/storage-provisioner] Handler 'handle_legacy_pod' succeeded.
```

6. Como a nuestro operador solo le intersan los `Pod` asociados a Legacy,
   filtraremos los eventos que esten asociados al label `secret-handshake`

```python
import kopf

@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT})
async def handle_legacy_pod(event, status, namespace, name, logger, **_):
  logger.info(f"Found {name} in {namespace} namespace")
```

7. Si ejecutamos nuestro operador en "modo dev" una vez más, veremos que ya no
   aparecen más logs, porque aún no está desplegado ningún `Pod` con el label
   `secret-handshake`

```shell
$ kopf run --all-namespaces legacy-automation.py
[... logs de inicialización ...]
```

8. Como lo más importante para inicializar Legacy es determinar su IP,
   filtraremos una vez más los eventos enfocándonos en obtener esta información
   de un `Pod` que ya esté en ejecución y utilice la label `secret-handshake`

```python
import kopf

@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT},)
async def handle_legacy_pod(event, status, namespace, name, logger, **_):
    if status.get('phase') != 'Running' or event.get('type') == 'DELETED':
        return

    podIP = status.get('podIP', '')
    if not podIP:
        return

    logger.info(f"=== Found {name} in {namespace} namespace w/ {podIP} address")

    # TODO: Add the automation logic here
```

1. Verificaremos una última vez en "modo dev" antes de agregar la lógica del
   "apreton de manos secreto". Para ello crearemos un `Pod` de pruebas primero y
   luego correremos nuestro operador.

```shell
$ kubectl run dummy-app --image busybox --labels secret-handshake=true --command -- sleep 90s
$ kopf run --all-namespaces legacy-automation.py
[... logs de inicialización ...]
[INFO    ] [default/dummy-app] === Found dummy-app in default namespace w/ 10.244.0.24 address
[INFO    ] [default/dummy-app] Handler 'handle_legacy_pod' succeeded.
```

10. _Opcional_. Elimina el `Pod` que creamos anteriormente para no confundir
    futuras pruebas

```shell
kubectl delete pod dummy-app
```

11. Ahora agregamos toda la lógica de inicialización del `Pod`. El resultado
    final debe verse similar a lo que encuentras en `kopf-operator/operator.py`

<details>

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

```

</details>

12. Construiremos la imagen de la solución final. (Si prefieres utilzar tu
    versión debes cambiar el argumento a `./operator`)

```shell
../_hack/build.sh ./kopf-operator
```

13. "Instalamos" el operador dentro del clúster

```shell
$ kubectl apply -f manifests/kopf-operator-install.yaml
namespace/legacy-operator-system created
serviceaccount/kopf-legacy-operator-sa created
clusterrole.rbac.authorization.k8s.io/kopf-legacy-operator-cluster-role created
clusterrolebinding.rbac.authorization.k8s.io/kopf-legacy-operator-cluster-role-binding created
deployment.apps/legacy-operator created
```

14. Verificamos que el operador esté corriendo sin problemas

```shell
$ kubectl get deployment -n legacy-operator-system
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
legacy-operator   1/1     1            1           12s
$ kubectl logs -n legacy-operator-system -l application=legacy-operator
[INFO    ] Initial authentication has been initiated.
[INFO    ] Activity 'login_via_client' succeeded.
[INFO    ] Initial authentication has finished.
```

### Haciendo la prueba

Una vez nuestro operador está corriendo, necesitamos desplegar a Legacy

- Construimos la imagen de Legacy

```shell
../_hack/build.sh ./mock-app
```

- Aplicamos el manifest de Legacy en Kubernetes

```shell
$ kubectl apply -f manifests/legacy-mock.yaml
namespace/legacy-automation-demo created
deployment.apps/legacy-mock created
service/legacy-mock-service created
ingress.networking.k8s.io/legacy-mock-ingress created
```

- Nos aseguramos que la aplicación esté corriendo

```shell
$ kubectl get deployment -n legacy-automation-demo
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
legacy-mock       1/1     1            1           1m33s
```

- _Opcional_ Si hiciste el proceso manual anteriormente, seguramente tienes un
  Pod de Legacy activo así que lo eliminaremos (no te preocupes, el Deployment
  se encargará de crear uno nuevo)

```shell
kubectl delete pod --selector app=legacy-mock
```

- Ahora revisamos si el **nuevo** Pod de Legacy ya está inicializado

```shell
$ curl -i http://localhost/
HTTP/1.1 503 Service Unavailable
Content-Type: text/plain; charset=utf-8
Content-Length: 14
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

Uninitialized
```

¡Un momento el nuevo Pod de Legacy no está inicializado! ¿Qué ha sucedido? 🤔.

Si te recuerdas nuestro operador utiliza el label `secret-handshake` para
determinar cual `Pod` inicializar. Para arreglar esto haremos lo siguiente:

- Descomentamos de `manifests/legacy-mock.yaml` el label `secret-handshake` y
  aplicamos este nuevo cambio al cluster

```shell
# sed 's/# //g' manifests/legacy-mock.yaml
$ kubectl apply -f manifests/legacy-mock.yaml
namespace/legacy-automation-demo unchanged
deployment.apps/legacy-mock configured
service/legacy-mock-service unchanged
ingress.networking.k8s.io/legacy-mock-ingress unchanged
```

- Verificamos si esta vez el nuevo Pod de Legacy está inicializado

```shell
$ curl -i http://localhost/
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 10
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

It works!
```

<!--
```shell
kubectl run -n legacy-automation-demo --rm --image-pull-policy IfNotPresent \
    --image curlimages/curl test --restart Never -i \
    -- curl -si http://legacy-mock-service.legacy-automation-demo.svc.cluster.local:3000/

Content-Type: text/plain; charset=utf-8
Content-Length: 10
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

It works!
```
 -->

¡Funciona! Ahora ya no tenemos que estar persiguiendo los Pod de Legacy cada vez
que arranquen.

## Otros detalles

Hay algunos detalles que tuvimos que saltarnos para mantener la narrativa del
ejercicio corta, pero son igual de importantes así que los compartiremos acá.

### ¿Por qué la solución final tiene un filtro más sobre los annotations?

A pesar de los demás filtros que colocamos previamente, el evento que buscamos
puede suceder más de una vez y por ende el ciclo de inicialización puede ser
lanzado varias veces aunque no sea necesario. Por ello utilizamos una
[_annotation_](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) (anotación), para determinar que Pods fueron previamente
inicializados para ignorarlos más adelante.

Puedes encontrar como agregamos este _annotation_ al Pod dentro de la función
`secret_handshake` que sucede al final del ciclo de inicialización.

```python
        kubernetes.config.load_incluster_config()
        api = kubernetes.client.CoreV1Api()
        api.patch_namespaced_pod(name=name,namespace=namespace, body=[{
            'op': 'add',
            'path': '/metadata/annotations',
            'value': {
                'secret-handshake-done': 'true'
            }
        }])
```

### ¿El operador tiene acceso todos los recursos del clúster?

No, el operador tiene un acceso restringido al Kubernetes API.

Dentro del archivo `manifests/kopf-operator-install.yaml` podrás encontrar todos
los permisos le damos a nuestro operador por medio de un
[ClusterRole](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole)
+
[ServiceAccount](https://kubernetes.io/docs/concepts/security/service-accounts/).

### ¿Le agregarías algo más a la solución final?

Si, de nuevo por fines didácticos se priorizó que el código sea fácil de leer,
por ello evitamos agregar un par de cosas más. Como por ejemplo:

- Podríamos mover el filtrado de eventos a una
  [expresión lambda](https://docs.python.org/3/tutorial/controlflow.html#lambda-expressions)
  en el parametro `when` de la anotación `@kopf.on.event`. Viendose algo así:

```python
@kopf.on.event('pods',
               labels={'secret-handshake': kopf.PRESENT},
               annotations={'secret-handshake-done': kopf.ABSENT},
               when= lambda status, event, **_: event.get('type') != 'DELETED' and status.get('phase') == 'Running')
async def handle_legacy_pod(event, status, namespace, name, logger, **_):
    podIP = status.get('podIP', '')
    if not podIP:
        return

    logger.info(f"=== Found {name} in {namespace} namespace w/ {podIP} address")

    asyncio.create_task(secret_handshake(podIP, namespace, name, logger))
```

- Podríamos mover el requerimiento del label `secrect-handshake` a nivel del
  Deployment y que este se "pase" al Pod cuando se crea. Pero eso requeriría:
    - Que el operador tenga accesso a `apps/Deployment` en los verbos `get`,
      `list` y `patch`
    - Agregar un nueva función con la anotación `@kopf.on.create`

```python
@kopf.on.create('deployments', labels={'secret-handshake': kopf.PRESENT})
async def handle_legacy_deployment():
    # Lógica de agregar el label a los Pods dentro de deployment.spec.template.metadata.labels
```

### ¿Debería utilizar kopf para mis operadores o recomiendas otros frameworks?

Depende. Cuando comparas herramientas no hay mejor o peor, sino cuándo y cómo
las utilizas.

El framework kopf es una excelente herramienta para crear pruebas de concepto
rápidamente, aprender experimentando y compartir ideas en un lenguaje de
programación amigable al lector. Aún así, en proyectos que tendrán varios
contribuidores muchas veces es preferible optar por frameworks con "opiniones
fuertes" en cuanto mejores prácticas y cómo deberían ser utilizados.

Una alternativa que puedes explorar es [kubebuilder](https://kubebuilder.io/).
