# Hablemos de autómatas, blogs y Operadores de Kubernetes

Bienvenido a la introducción práctica a Operadores de Kubernetes. Si tuviste la
oportunidad de estar en la plática podrás identificar fácilmente los ejemplos
que se encuentran en este repositorio, en caso contrario no te preocupes trataré
de explicar lo más que se pueda por medio de la documentación interna.

En la raíz de este repositorio encontrarás los siguientes directorios, los
cuales tienen su propio enfoque:

- **legacy**. Contiene los componentes descritos en la historia *legacy*. Los
  cuales son una mock-app (nuestro "legacy"), el código fuente de un operador
  utilizando [kopf](https://github.com/nolar/kopf) y los manifestos que nos
  permiten provisionarlos.
- **\_hack**. Acá están todos los scripts que nos ayudaran a hacer fácilmente
  setup, build y publish de los artefactos necesarios para las pláticas
  descritas anteriormente.

## Requerimientos

Los ejercicios descritos en este repositorio necesita las siguientes
herramientas para poder funcionar:

- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [docker](https://www.docker.com/products/docker-desktop/)

En caso desees experimentar con el código fuente de las aplicaciones entonces necesitarás:

- Python 3.9+
- Golang 1.18+

## Otros recursos

El contenido de este repositorio se basa en los siguietes materiales que
elaboran a más profundidad:

- Kubernetes Operators Documentation <https://kubernetes.io/docs/concepts/extend-kubernetes/operator/>
- CNCF Operator Whitepaper <https://tag-app-delivery.cncf.io/whitepapers/operator/>
- KOPF Documentation <https://kopf.readthedocs.io/en/stable/>
- Operator SDK Documentation <https://sdk.operatorframework.io/docs/>
- Kubernetes Operators Explained <https://www.youtube.com/watch?v=i9V4oCa5f9I>
- Building operators with the Operator SDK <https://www.youtube.com/watch?v=5XZZxhwb_xs>
- Go! Bash Operator <https://www.youtube.com/watch?v=we0s4ETUBLc>
