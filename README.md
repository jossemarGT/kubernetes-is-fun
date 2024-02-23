# Hablemos de autómatas y Operadores de Kubernetes

Bienvenido a la introducción práctica a Operadores de Kubernetes. Si tuviste la
oportunidad de estar en la plática podrás identificar fácilmente los ejemplos
que se encuentran en este repositorio.

En la raíz de este repositorio encontrarás los siguientes directorios:

- **legacy**. Contiene el código fuente de nuestra aplicación "Legacy", el
  operador que maneja su despliegue y los manifestos que nos permiten
  provisionarlos.
- **\_hack**. Acá están todos los scripts que nos ayudaran a hacer fácilmente
  inicializar y construir los artefactos que encuentras en el repositorio.

## Requerimientos

Los ejercicios descritos en este repositorio necesita las siguientes
herramientas para poder funcionar:

- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [docker](https://www.docker.com/products/docker-desktop/)

En caso desees experimentar con el código fuente de las aplicaciones entonces necesitarás:

- [Python](https://www.python.org/downloads/) 3.9+
- [Golang](https://go.dev/doc/install) 1.18+

## Otros recursos

El enfoque de este repositorio es dar una introducción práctica en Español a
operadores, pero si buscas material que va a más profundidad puedes utilizar:

- Kubernetes Operators Documentation <https://kubernetes.io/docs/concepts/extend-kubernetes/operator/>
- CNCF Operator Whitepaper <https://tag-app-delivery.cncf.io/whitepapers/operator/>
- KOPF Documentation <https://kopf.readthedocs.io/en/stable/>
- Operator SDK Documentation <https://sdk.operatorframework.io/docs/>
- Kubernetes Operators Explained <https://www.youtube.com/watch?v=i9V4oCa5f9I>
- Building operators with the Operator SDK <https://www.youtube.com/watch?v=5XZZxhwb_xs>
