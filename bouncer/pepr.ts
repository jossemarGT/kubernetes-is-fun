import { PeprModule } from "pepr";
import cfg from "./package.json";
import { Bouncer } from "./capabilities/bouncer";

new PeprModule(cfg, [Bouncer]);
