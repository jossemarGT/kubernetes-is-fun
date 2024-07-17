# Kubernetes es divertido

Por alguna razón, muchos hemos llegado a pensar que Kubernetes no es divertido
para desarrolladores, cuando en realidad hay muchas maneras de
[extenderlo](https://kubernetes.io/docs/concepts/extend-kubernetes/) en el
lenguaje de programación que más nos acomode.

Exploremos como podemos extender el comportamiento de Kubernetes.

<!-- Tal vez hasta lograremos hacer que Kubernetes salte ;) -->

## ¿Dónde inicio?

Este repositorio está conformado por enunciados de ejercicios sencillos y la
solución final. Si deseas ir paso a paso, puedes seguir la conversación en la
grabación (PENDIENTE) o bien guiarte por el `README.md` en cáda directorio y
explorar por tu cuenta.

El orden recomendado en base a su complejidad es el siguiente:

- `bouncer/` Extiende el API de Kubernetes creando tus propios Admission
  Webhooks con pepr.
- `less/` Crea un Kubernetes Controller utilizando helm y el Operator SDK.
- `legacy/` Crea un Operador para automatizar el provisionamiento de una
  aplicación Legacy.

Pero eres bienvenido a explorar de la manera que más te funcione 😉.

### Requerimientos

Los ejercicios anteriormente mencionados están intencionalmente escritos en
múltiples lenguajes, valiendose de diferentes herramientas pero como mínimo
necesitarás tener instalado:

- [docker](https://www.docker.com/products/docker-desktop/)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)

Cada solución indica que lenguaje de programación utilizaremos, pero si te
quieres adelantar puedes instalar:

- [Python](https://www.python.org/downloads/) 3.9+
- [Golang](https://go.dev/doc/install) 1.18+
- [Node.js](https://github.com/nvm-sh/nvm) v20+

### Levantando nuestro clúster local de pruebas

Para realizar las soluciones a los ejercicios en este repositorio no necesitas
crear un clúster de Kubernetes en a la nube. Si ya tienes instalado `minikube`
solo necesitas hacer lo siguiente:

```shell
../_hack/setup.sh
```

Toda la configuración para conectarte al clúster local ya ha sido generada,
ahora ya puedes utilizar comandos de `kubectl` sin problema. Como por ejemplo:

```shell
$ kubectl get nodes
NAME      STATUS   ROLES           AGE     VERSION
k8s-fun   Ready    control-plane   1h      v1.30.2
```

Cuando termines o desees comenzar desde cero, puedes destruir el clúster con:

```shell
../_hack/clean-up.sh
```

**Importante**: Al crear nuestro clúster, `minikube` seleccionará el mejor
backend para tu entorno local. En algunos casos seleccionará `docker`, lo cual
implica que necesitas correr el siguiente comando en una nueva terminal para
poder acceder directamente a los endpoints de las aplicaciones de ejemplo.

```shell
minikube tunnel
```

En caso no sepas que backend selección `minikube` por ti puedes utilizar el
siguiente comando:

```shell
$ minikube profile list
|------------|-----------|---------|--------------|------|---------|---------|-------|----------------|--------------------|
|  Profile   | VM Driver | Runtime |      IP      | Port | Version | Status  | Nodes | Active Profile | Active Kubecontext |
|------------|-----------|---------|--------------|------|---------|---------|-------|----------------|--------------------|
| k8s-is-fun | docker    | docker  | 192.168.58.2 | 8443 | v1.30.2 | Running |     1 | *              | *                  |
|------------|-----------|---------|--------------|------|---------|---------|-------|----------------|--------------------|
```

## Quiero aprender más

Si terminaste con los ejercicios o tu gustaría leer material que va a más
profundidad puedes utilizar:

- Extending Kubernetes <https://kubernetes.io/docs/concepts/extend-kubernetes/>
- pepr documentation <https://docs.pepr.dev/>
- Operator SDK documentation <https://sdk.operatorframework.io/docs/>
- KOPF documentation <https://kopf.readthedocs.io/en/stable/>
- kubebuilder documentation <https://book.kubebuilder.io/>
- kube-rs documentation <https://kube.rs/getting-started/>
- Kubernetes RBAC <https://kubernetes.io/docs/reference/access-authn-authz/rbac>
- CNCF Operator Whitepaper <https://tag-app-delivery.cncf.io/whitepapers/operator/>
