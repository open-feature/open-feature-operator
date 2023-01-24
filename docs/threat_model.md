# Threat Model

Threat modeling is a structured approach of identifying and prioritizing potential threats to a system, and determining the value that potential mitigations would have in reducing or neutralizing those threats.

Using [OWASP's threat modeling cheat sheet for guidance](https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html).

## Application Entry Points

The interfaces through which potential attackers can interact with the application or supply them with data.
The diagram below models these interfaces, the sections that follow describe them further.

<img src="../images/ofo_threat_model.png" alt="Diagram of OFO threat model">

### Via the application to which OpenFeature Operator (OFO) sidecars flagd

OFO appends flagd as a sidecar container to any pod spec application with valid annotations. The application is then able to evaluate flags by calling flagd (typically via an sdk and flag provider). While a caller of the application could be from a trusted entity (internal infrastructure), it is more prudent to presume the agent to be untrusted.

### Via the Kubernetes API server

OFO listens to webhooks from the Kubernetes control plane (its API server), specifically it handles mutations of pods.
It parses the mutated pod's annotations to determine which CRD to retrieve flagd's flag configuration from, mutating flagd's internal state if necessary.


