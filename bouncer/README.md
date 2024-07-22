# Bouncer (Admission Webhook)

¿Qué tal si pudieramos evitar que lleguen malas configuraciones a nuestro
clúster de Kuberntes?

## El concepto

Kubernetes nos permiete extender su API con los
[Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
los cuales reciben las peticiones de creación, actualización o eliminación de
recursos antes de ser procesadas.

## El ejercicio a resolver

El "Senior" de nuestro equipo comenta que a varios compañeros se les olvida
colocar la etiqueta `equipo` en las cargas de trabajo y configuraciones que
aplican a Kubernetes. Esta _label_ es importante porque ayuda en varios procesos
internos de la empresa.

Nosotros como desarrolladores sabemos que Kubernetes tiene un API y tal vez
podríamos crear algo que funcione como _middleware_ que **valide** cada uno de
los objetos que llegan a el. Así podríamos filtrar solo los objetos que cumplen
con las normas.

## La solución

Utilizaremos a [pepr](https://pepr.dev/) para escribir fácilmente nuestros
Admission Webhooks en [TypeScript](https://www.typescriptlang.org/).

1. Inicializamos nuestra aplicación

```shell
npx pepr init
```

2. Creamos un nuevo _capability_ que será nuestro WebHook llamado `bouncer`.

```javascript
// capabilities/bouncer.ts

import { Capability, a } from "pepr";

export const Bouncer = new Capability({
  name: "bouncer",
  description:
    "Verificando que solo los recursos cool ingresen al Kubernetes API",
});

const { When } = Bouncer;
```

3. Por cada tipo de _Kubernetes Object_ (en este caso un `ConfigMap`) que
   querramos validar haremos lo siguiente

```javascript
When(a.ConfigMap)
  .IsCreatedOrUpdated()
  .Validate(request => {
    if (request.HasLabel("equipo")) {
      return request.Approve();
    }

    return request.Deny(`${request.Raw.kind} debe tener label 'equipo'`);
  });
```

4. ¿Que tal si agregamos al usuario que lo creó o modificó?

```javascript
When(a.ConfigMap)
  .IsCreatedOrUpdated()
  .Mutate(request => {
    const { username } = request.Request.userInfo;

    request.Merge({
      metadata: {
        labels: {
          bouncer: "estuvo-aqui",
        },
        annotations: {
          "creado-por": username,
        },
      },
    });
  })
  .Validate(request => {
    // ... código de validación de la sección anterior...
  });
```

5. "Montamos" nuestro WebHook al entrypoint de nuestra aplicación `pepr.ts`

```javascript
import { PeprModule } from "pepr";
import cfg from "./package.json";
import { Bouncer } from "./capabilities/bouncer";

new PeprModule(cfg, [Bouncer]);
```

6. Construimos nuestra aplicación

```shell
npx pepr build
```

7. La desplegamos en nuestro clúster

```shell
npx pepr deploy
# O de manera alterantiva podemos usar el manifest directamente
# kubectl apply -f dist/pepr-*.yaml
```

8. Verificamos que esté lista

```shell
$ kubectl get deploy -n pepr-system
NAME                                        READY   UP-TO-DATE   AVAILABLE   AGE
pepr-4b8c9479-f7c6-5f20-894c-37353c5deece   2/2     2            2           61s
```

### Haciendo la prueba

Apliquemos un `ConfigMap` que **no** cumple con los requerimientos

```shell
$ kubectl apply -f manifests/bouncer.samples.yaml
namespace/bouncer-demo created
Error from server: error when creating "manifests/bouncer.samples.yaml": admission webhook [...] 
denied the request: ConfigMap debe tener label 'equipo'
```

Descomentamos los labels que hacen falta y volvemos a intentar

```shell
# sed 's/# //g' manifests/bouncer.samples.yaml
$ kubectl apply -f manifests/bouncer.samples.yaml
namespace/bouncer-demo unchanged
configmap/ejemplo created
```

¡Funciona! Ahora podemos agregar los demás objetos que queremos validar.
