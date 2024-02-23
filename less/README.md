# Less is more

Kubernetes ya nos ofrece bastantes abstracciones, pero a veces manejar menos de
ellas es mejor.

## El problema

¡Lo logramos! En nuestra empresa ya usan Kubernetes, pero por alguna razón el
arquitecto anterior quiso que se utilizara para practicamente todo. Lo cual
rápidamente mostró que no era la mejor idea cuando llegan nuevas personas que
trabajaran temporalmente con nosotros.

Entre tantas cosas que cubrimos en nuestra empresa está publicar "la frase del
día" (FDD) que debe ser cambiada cada 24hrs y no se puede automatizar por reglas
de negocio 🤷. Entonces, cada mañana los primero que tiene que hacer nuestros
becario es:

1. Conseguir la frase del día (FDD)
2. Abrir "el yaml" de la aplicación para FDD
3. Buscar y actualizar la línea donde done está escrita la FDD anterior
4. Guardar y hacer `kubectl apply -f` del nuevo yaml
5. ???
6. ¡Listo!

¿Qué puede salir mal si todo es sencillo? Lamentablemente muchas cosas como:

1. Un typo en la definición del yaml
2. El usuario del becario necesita permisos para modificar los tipos definidos
   en el yaml
3. El usuario del becario tácitamente tiene acceso a **editar** otros objectos
   dentro del cluster
4. ...
5. Mejor dejésmolo allí 😅

### Lo que el becario ve

```sh
kubectl apply -f app-fdd.yaml
```

## Trabajemos en el laboratorio local

Lo que verenmos en el local es:

1. Provisionar la FDD como el becario
2. Crear el helm chart en base al yaml
3. Generar Operador en base al helm chart


### Comandos desarrollo que debes recordar

[Instalar Operator SDK cli](https://sdk.operatorframework.io/docs/installation/)

```sh
$ export ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
$ export OS=$(uname | awk '{print tolower($0)}')
$ export OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.33.0
$ curl -LO ${OPERATOR_SDK_DL_URL}/operator-sdk_${OS}_${ARCH}
$ chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk

```

Generar el Operador a partir de un helm chart

```sh
# dentro de un nuevo directorio
operator-sdk init \
  --plugins helm \
  --domain lets-talk-about-operators.local \
  --group apps \
  --version v1alpha1 \
  --kind FraseDelDia \
  --helm-chart ../helm-charts/fdd \
  --project-name less-helm-operator
```

Construir el operador

```sh
# Dentro del directorio del operador
make -e IMG=lets-talk-about-operators/less-helm-operator:0.1.0 docker-build
```

Correr el Operador en modo desarrollo

```sh
# Dentro del directorio del operador
make -e IMG=lets-talk-about-operators/less-helm-operator:0.1.0 install run
```

Preparar para producción el Operador y sus CRDs

```sh
# Dentro del directorio del operador
cd config/manager
kustomize edit set image controller=lets-talk-about-operators/less-helm-operator:0.1.0
kustomize build config/default >manifests/install-fdd-helm-operator.yaml
kustomize build config/crd >manifest/install-fdd-helm-operator-crds.yaml
```
