# Delete unused dev volumes within a development namespace

> This is an experiment and Okteto does not officially support it.

This is a simple example on how Okteto's public API could be combined with Okteto CLI and other tools like kubectl to create automatic tasks to delete unused dev volumes within development namespaces.

> You might want to add some filter to the script to delete the volumes in a more controlled way. Dev volumes contain the code synchronized from local machines, so deleting them might cause slower development experience the next time a developer starts the development environment.

- Create an [Okteto Admin Token](https://www.okteto.com/docs/admin/dashboard/#admin-access-tokens)

- Export the token to a local variable:

```bash
export OKTETO_ADMIN_TOKEN=<<your-token>>
```

- Create a namespace, and, via the admin section, mark it as [Keep awake](https://www.okteto.com/docs/admin/dashboard/#namespaces)

- Export the namespace name to a local variable:

```bash
export NAMESPACE=<<your-namespace>>
```

- Create a local variable to define the cronjob schedule:

```bash
export JOB_SCHEDULE="0 20 * * *"
```

For example, 0 0 13 * 5 states that the task must be started every Friday at midnight, as well as on the 13th of each month at midnight.

- Run the following command to create the cronjob:

```bash
okteto deploy -n ${NAMESPACE} --var OKTETO_ADMIN_TOKEN=${OKTETO_ADMIN_TOKEN} --var JOB_SCHEDULE=${JOB_SCHEDULE}
```

## Force the execution of the job

To force the execution of the job, run the following commands:

```bash
okteto kubeconfig
kubectl -n ${NAMESPACE} create job --from=cronjob/delete-dev-volumes delete-dev-volumes-$(date +%s)
```