// capabilities/bouncer.ts

import { Capability, a } from "pepr";

export const Bouncer = new Capability({
  name: "bouncer",
  description:
    "Verificando que solo los recursos cool ingresen al Kubernetes API",
});

const { When } = Bouncer;

When(a.ConfigMap)
  .IsCreatedOrUpdated()
  .InNamespace("bouncer-demo")
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
    if (request.HasLabel("equipo")) {
      return request.Approve();
    }

    return request.Deny(`${request.Raw.kind} debe tener label 'equipo'`);
  });
