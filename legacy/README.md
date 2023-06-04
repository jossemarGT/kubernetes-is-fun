# Legacy

> Tu equipo ha logrado migrar la mayoría de las aplicaciones a Kubernetes. Sin
> embargo, aún queda una sola aplicación que pocos saben de cómo apareció o en
> que lenguaje está escrita. 
>
> Todos la llaman ¡Legacy!
>
> Legacy ya corre como contenedor pero aún necesita varios pasos manuales para
> que arranque. 
> 
> ¡Ayudanos a escapar a la realidad que Legacy ha creado!

Tal como conversamos en nuesra plática en vivo, necesitamos automatizar los
pasos de inicialización de una aplicación "Legacy". En este punto ya descartamos
las siguientes opciones:

- **Mantener el proceso manual**. Esto es lo que nuestros compañeros han hecho al
  día de hoy, pero no es escalable y necesitamos cambiarlo.
- **Modificar a Legacy**. Lamentablemente el último que entendía el lenguaje en
  que está escrita Legacy ya no está con nosotros, así que esto es muy
  complicado de lograrlo.
- **Delegar proceso de despliegue a una solución de CI**. Esto puede funcionar,
  pero estaríamos agregando un segundo punto de falla en nuestro sistema; además
  los servidores de CI no son la mejor opción para mantener valores sensibles.
- **Forzar Kubernetes**. Cualquier otra "solución creativa" que logre forzar a
  Kubernetes a funcionar fuera de como fue diseñado. Puede funcionar, pero a la
  larga vueve los sistemas más dificiles de mantener.

Así que tomaremos la ruta de crear nuestro propio Operador de Kubernetes para
automatizar este proceso desde dentro del cluster y de manera autónoma.

- [La historia continúa](#la-historia-continúa)
  - [Hagamos el despliegue manual de Legacy](#hagamos-el-despliegue-manual-de-legacy)
  - [Automatizemos con operadores](#automatizemos-con-operadores)
  - [Somos libres de Legacy y ¿ahora qué?](#somos-libres-de-legacy-y-ahora-qué)
  - [Otros detalles de interés](#otros-detalles-de-interés)
- [Configurando laboratorio local](#configurando-laboratorio-local)
- [Acerca del código de nuestra historia](#acerca-del-código-de-nuestra-historia)

## La historia continúa

Legacy es una aplicación que expone un
[Restful API](https://aws.amazon.com/es/what-is/restful-api/), la cual necesita
ser inicializada previo a que sirva cualquier petición pública. Legacy ya es
capáz de correr como contenedor, pero que requiera un proceso de inicialización
necesita de atención constante de nuestro equipo.

El proceso de inicialización consiste en acceder a un
[http endpoint]((https://www.cloudflare.com/es-es/learning/security/api/what-is-api-endpoint/))
interno el cual expone la mitad de una cadena secreta. Esta cadena debe ser
tomada por un agente externo y unirla con la otra mitad; una vez unidas debe ser
enviada como solo una cadena a otro endpoint interno. Cuando este proceso se
hace correctamente Legacy permitirá que se accedan a los endpoints públicos.

Los endpoints que Legacy expone son:

- `GET /health` siempre responde 200 toda vez Legacy esté corriendo.
- `GET /internal/key` toda vez Legacy no esté inicializada responderá con la
  *mitad* de la cadena secreta.
- `POST /internal/secret` acepta la cadena secreta completa. En caso de ser la
  cadena correcta responderá 204 y desbloqueará los endpoints públicos.
- `/` representa a todos los endpoints públicos y retornará error mientras
  Legacy no esté inicializada.

### Hagamos el despliegue manual de Legacy

> ℹ️ **Recuerda**
> Puedes hacer esto desde tu [laboratorio local](#configurando-laboratorio-local).

Legacy ya puede ser desplegada como
[Deployment](https://kubernetes.io/es/docs/concepts/workloads/controllers/deployment/),
con su respectivos
[Service](https://kubernetes.io/es/docs/concepts/services-networking/service/) e
[Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/). Aún
así requiere tenemos que inicializarla manualmente como lo describimos en la
sección anterior, en código se vería de la siguiente manera:

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

$ curl -i http://localhost/health
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Content-Length: 16
Connection: keep-alive
X-App-Name: legacy-mock
X-App-Version: 0.1.0

{"status":"ok"}
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

- Enviamos la cadena compelta de vuelta a Legacy

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

Crear un operador de Kubernetes **desde cero** requiere tener un conocimiento de
detalles como la arquitectura de los
[controllers](https://kubernetes.io/docs/concepts/architecture/controller/) y
como funciona el
[API de Kubernetes](https://kubernetes.io/es/docs/concepts/overview/kubernetes-api/),
por ello usaremos la ayuda de un framework para simplificar el desarrollo del
nuestro.

En este ejemplo usaremos el
[Kubernetes Operator Pythonic Framework](https://github.com/nolar/kopf) o
*KOPF*, dado que Python es sencillo de comprender para la mayoría de personas.
Aún así debes saber que existen muchos más frameworks y toolkits que se pueden
utilizar en su lugar dependiendo a las necesidades que buscas suplir. El código
fuente de nuestro operador lo puedes encontrar en el directorio `kopf-operator`.

Es más sencillo explicar como funciona una vez lo veamos en acción y para ello
debemos hacer lo siguiente:

- "Instalamos" nuestro operador

```sh
$ cd legacy
$ kubectl apply -f manifests/kopf-operator-install.yaml
```

- Verificamos que esté corriendo

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

- Ahora revisamos si el nuevo pod de Legacy ya está inicializado

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

¡Un momento! El nuevo pod de Legacy no está inicializado ¿qué ha pasado?. No
preocupes, el Operador si notó el nuevo pod sin embargo decidió ignorarlo porque
no traía consigo el label `secret-handshake`. Hemos agregado esta condicional
extra en el operador para que pueda diferenciar los pods de Legacy que deben ser
inicializados de los que no. Para arreglar esto haremos lo siguiente:

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

¡Listo! De ahora en adelante ya no tendremos más porque estar al tanto de cuando
un nuevo Pod de Legacy es creado.

### Somos libres de Legacy y ¿ahora qué?

El código de nuestro `legacy-operator` es bastante simple por fines didácticos.
Aún así podemos considerar las siguientes mejoras:

- Complementar la solución con un
  [Custom Resource Definition](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).
  propio. De esta manera se aisla declarativamente que aplicaciones son Legacy
  de las demás y también nos permite exponer explicitamente otras
  configuraciones como el puerto que usará esta misma.
- De no interesarnos mantener un Custom Resource Definition, tambien podemos
  explorar el monitorear el label `secret-handshake` a nivel del Deployment en
  lugar de Pod. De esta manera podemos determinar el puerto de Legacy al
  inspeccionar el Service asociado a este.

Lo bueno de frameworks como [Operator SDK](https://sdk.operatorframework.io/) o
[KOPF](https://kopf.readthedocs.io/en/stable/) es que nos facilitan experimentar
con las posibilidades.

### Otros detalles de interés

Algo que no mencionamos durante la historia es que los operadores necesitan
permisos para poder interactuar con el
[API de Kubernetes](https://kubernetes.io/es/docs/concepts/overview/kubernetes-api/).
Esta asignación de permisos los puedes observar al inicio de
`manifests/kopf-operator-install.yaml` y a todo eso se le denomina
[Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac)
que podemos abordar en otra ocasión.

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

**NOTA**: Cuando crees el cluster de nuestro laboratorio, `minikube`
seleccionará el mejor backend para tu entorno local. Aún así, **en caso** el
backend seleccionado sea `docker` o bien utilizas OSX necesitas hacer lo
siguiente en una *nueva terminal*:

```sh
minikube tunnel
```

## Acerca del código de nuestra historia

La mejor parte de esta historia es poder revisar el código fuente y experimentar con el mismo. Nuestro ejemplo está divido en tres directorios principales:

- `mock-app` Contiene una pequeña aplicación escrita en Go simula el
  comportamiento de Legacy. Cabe resaltar que su código es derivado de
  [hashicorp/http-echo](https://github.com/hashicorp/http-echo).

- `kopf-operator` Contiene la lógica de nuestro operado y está escrito en Python
  3. Lo más interesante es que gracias a KOPF pudimos crear un operado en unas
  pocas líneas de código. contiene la lógica de nuestro operador escrito en
  Python 3 utiliza el Kubernetes Pythonic Operator Framework (kopf)

- `manifests` Contiene todos los Kubernetes Manifest o bien archivos YAML que
  provisionan a Legacy y su operador.
