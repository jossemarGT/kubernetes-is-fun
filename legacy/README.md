# Legacy

Tu equipo pensaba que solo era de empaquetar una aplicación en un contenedor y
que luego todo sería paz despues de eso. Pronto descubriríamos que "Legacy"
tenía más sorpresas por dentro.

Nuestro equipo ya ha intentando suficientes soluciones creativas sin suerte así
que ahora es tiempo de salvar el día con Kubernetes y nuestro propio
[operador](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

- [La historia continúa](#la-historia-continúa)
  - [Haciendo un despliegue manual de Legacy](#haciendo-un-despliegue-manual-de-legacy)
  - [Automatizemos con operadores](#automatizemos-con-operadores)
  - [Somos libres de Legacy y ¿ahora qué?](#somos-libres-de-legacy-y-ahora-qué)
  - [Otros detalles de interés](#otros-detalles-de-interés)
- [Configurando laboratorio local](#configurando-laboratorio-local)
- [Acerca del código de nuestra historia](#acerca-del-código-de-nuestra-historia)

## La historia continúa

Legacy es una aplicación que expone una
[Restful API](https://aws.amazon.com/es/what-is/restful-api/), la cuál no aceptará ninguna petición pública hasta que se inicialize manualmente.

El proceso de inicialización consiste en acceder a un
[http endpoint]((https://www.cloudflare.com/es-es/learning/security/api/what-is-api-endpoint/))
interno para obtener la mitad de una cadena secreta, nuestro equipo la completa
con la otra mitada y luego la envían de vuelta a la aplicación. Hasta que este
proceso se haga correctamente Legacy será completamente funciona.

Los endpoints que Legacy expone son:

- `GET /health` siempre responde `200` toda vez Legacy esté corriendo.
- `GET /internal/key` responde con la *mitad* de la cadena secreta, mientras la
  aplicación no esté inicializada.
- `POST /internal/secret` acepta la cadena secreta completa. En caso de ser la
  cadena correcta responderá `204` y desbloqueará los demás endpoints.
- `/` representa a los demás endpoints y retornará error `500` mientras
  la aplicación no esté inicializada.

### Haciendo un despliegue manual de Legacy

> ℹ️ **Recuerda**
> Puedes hacer esto desde tu [laboratorio local](#configurando-laboratorio-local).

Legacy ya puede ser desplegada como
[Deployment](https://kubernetes.io/es/docs/concepts/workloads/controllers/deployment/),
con su respectivos
[Service](https://kubernetes.io/es/docs/concepts/services-networking/service/) e
[Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/). Aún
así requiere que se inicialize manualmente como lo describimos anteriormente; en
código el proceso se vería de la siguiente manera:

- Aplicamos nuestro manifesto (el archivo yaml de Legacy)

```sh
$ cd legacy
$ kubectl apply -f manifests/legacy-mock.yaml
```
  
- Verificamos si Legacy está corriendo

```sh
$  kubectl get deployments
NAME          READY   UP-TO-DATE   AVAILABLE   AGE
legacy-mock   1/1     1            0           15s
```

```sh
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

```sh
$ curl http://localhost/internal/key
VGhpcyBpcyBAICQzY3IzdCBrZXk=
```

- Generamos la cadena completa de nuestro lado

```sh
$ curl -s http://localhost/internal/key | base64 -d | xargs -I {} echo '{}-acknowledge' | base64
VGhpcyBpcyBAICQzY3IzdCBrZXktYWNrbm93bGVkZ2UK
```

- Enviamos la cadena completa de vuelta a Legacy

```sh
$ curl -i -XPOST -d 'VGhpcyBpcyBAICQzY3IzdCBrZXktYWNrbm93bGVkZ2UK' http://localhost/internal/secret
HTTP/1.1 202 Accepted
Content-Type: text/plain; charset=utf-8
Content-Length: 9
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

Accepted
```

- Revisamos si Legacy ya está sirviendo los endpoints públicos

```sh
$ curl -i http://localhost/
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 16
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

It works!
```

### Automatizemos con operadores

> ℹ️ **Recuerda**
> Puedes hacer esto desde tu [laboratorio local](#configurando-laboratorio-local).

Crear un operador de Kubernetes **desde cero** requiere tener un basto
conocimiento acerca de
[controllers](https://kubernetes.io/docs/concepts/architecture/controller/) y
como funciona el
[API de Kubernetes](https://kubernetes.io/es/docs/concepts/overview/kubernetes-api/),
por eso es mejor ayudarse de un framework para simplificar el desarrollo de
cualquier idea que tengamos. En este ejemplo usaremos el
[Kubernetes Operator Pythonic Framework](https://github.com/nolar/kopf) ó
*KOPF*, dado que Python es sencillo de leer, comprender y explicar. Aún así
debes saber que existen muchos más frameworks y toolkits que se pueden utilizar
en su lugar dependiendo a las necesidades que buscas suplir. El código fuente de
nuestro operador lo puedes encontrar en el directorio `kopf-operator`.

Más adelante exploraremos el código, pero antes veamos como funciona la solución
al problema en caso de que algo salga mal. Para ello debemos seguir estos pasos:

- Primero debemos construir todos los artefactos que necesitamos tal como se
  describe en la sección de
  [configurando tu laboratorio local](#configurando-laboratorio-local).

- Nos aseguramos que nuestro legacy-mock esté corriendo

```sh
$ cd legacy
$ kubectl get Deployment legacy-mock
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
legacy-mock       1/1     1            1           1m33s
```

- "Instalamos" nuestro operador

```sh
$ cd legacy
$ kubectl apply -f manifests/kopf-operator-install.yaml
```

- Verificamos que el operador esté corriendo

```sh
$ kubectl get deployment --field-selector metadata.name=legacy-operator
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
legacy-operator   1/1     1            1           2m49s
```

- Para el momento de la verdad, eliminaremos cualquier pod de Legacy que este
  corriendo para dejar que nuestro operador lo elimine

```sh
kubectl delete pod --selector app=legacy-mock
```

- Ahora revisamos si el **nuevo** Pod de Legacy ya está inicializado

```sh
$ curl -i http://localhost/
HTTP/1.1 503 Service Unavailable
Content-Type: text/plain; charset=utf-8
Content-Length: 14
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

Uninitialized
```

¡Un momento! El nuevo Pod de Legacy no está inicializado ¿qué ha pasado?.

No te preocupes, el Operador si notó el nuevo pod sin embargo decidió ignorarlo
porque no traía consigo el label `secret-handshake`. Hemos agregado esta
condicional extra en el operador para que pueda diferenciar los pods de Legacy
que deben ser inicializados de los demás corriendo en el cluster.

Para arreglar esto haremos lo siguiente:

- Descomentamos de `manifests/legacy-mock.yaml` el label `secret-handshake`

```sh
sed -i 's/# secret-handshake.*/secret-handshake: "true"/' manifests/legacy-mock.yaml
```

- Aplicamos este nuevo cambio al cluster

```sh
$ kubectl apply -f manifests/legacy-mock.yaml
deployment.apps/legacy-mock configured
service/legacy-mock-service unchanged
ingress.networking.k8s.io/legacy-mock-ingress unchanged
```

- Verificamos si esta vez el nuevo Pod de Legacy está inicializado

```sh
$ curl -i http://localhost/
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 10
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

It works!
```

De ahora en adelante cada vez que un nuevo pod de Legacy arranque será
inicializado por nuestro Operador.

¡Legacy ha sido reducido al orden!

### Somos libres de Legacy y ¿ahora qué?

El código de nuestro `legacy-operator` es bastante simple por fines didácticos.
Aún así podemos considerar las siguientes mejoras:

- Para mantener la aplición sencilla podríamos mover el label `secret-handshake`
  que actualmente se define a nivel de Pod, hacia la definición de la carga de
  trabajo (Deployment). De esta podríamos identificar más fácilmente las
  configuraciones asociadas al Deployment, Service e Ingress que comparten los
  mismos selectors.
- De manera alternativa podríamos considerar crear un
  [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
  solamente para Legacy, por sus compartamientos únicos.

Lo bueno de frameworks como [Operator SDK](https://sdk.operatorframework.io/) ó
[KOPF](https://kopf.readthedocs.io/en/stable/) es que nos facilitan experimentar
con todas las alternativas que podamos imaginar.

### Otros detalles de interés

Algo que no mencionamos durante la historia es que los operadores necesitan
permisos para poder interactuar con el
[API de Kubernetes](https://kubernetes.io/es/docs/concepts/overview/kubernetes-api/).
Esta asignación de permisos los puedes observar al inicio de
`manifests/kopf-operator-install.yaml` y a todo eso se le denomina
[Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac)
que podemos abordar en otra ocasión 😉.

## Configurando laboratorio local

Para que puedas experimentar con el código de esta historia puedes levantar tu
laboratorio local de la siguiente manera:

- En una terminal cambia al directorio `legacy`

```sh
cd legacy
```

- Inicializa el laboratorio usando minikube.

```sh
../_hack/setup.sh
```

- Construye y carga las imagenes de nuestras aplicaciones (Legacy y el operador)

```sh
../_hack/build.sh ./mock-app
../_hack/build.sh ./kopf-operator
```

**NOTA**: Al crear nuestro cluster `minikube` seleccionará el mejor backend para
tu entorno local. Aún así, **en caso** el backend seleccionado sea `docker` (ó
utilizas OSX ) necesitas hacer lo siguiente en una *nueva terminal* para
interactuar directamente con Legacy:

```sh
minikube tunnel
```

## Acerca del código de nuestra historia

La mejor parte de esta historia es poder revisar el código fuente y experimentar
con el mismo. Nuestro ejemplo está divido en tres directorios principales:

- `mock-app` Contiene una pequeña aplicación escrita en Go simula el
  comportamiento de Legacy. Esta aplicación es derivada de
  [hashicorp/http-echo](https://github.com/hashicorp/http-echo).

- `kopf-operator` Contiene la lógica de nuestro operado, utilzando está escrito
  en Python 3 y el Kubernetes Pythonic Operator Framework (kopf).

- `manifests` Contiene todos los Kubernetes Manifest (archivos YAML) que definen
  como provisionar a Legacy y nuestro Operador.
