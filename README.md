# Delete unused PVCs within a development namespace

> This is an experiment and Okteto does not officially support it.

This is a simple example on how Okteto's public API could be combined with Okteto CLI and other tools like kubectl to create automatic tasks to delete unused PersistentVolumeClaims (PVCs) within development namespaces.

By default, the application only deletes unmounted **dev volumes** (PVCs with the label `dev.okteto.com=true`). You can optionally configure it to delete all unmounted PVCs. A PVC is considered "unused" if it is not currently mounted by any pod.

> **Warning:** When configured to delete all PVCs (`OKTETO_DEV_VOLUMES=false`), this tool will delete ALL unmounted PVCs, which may include data volumes. Make sure to review and test carefully before using in production environments.

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

- (Optional) Set environment variables for additional filtering:

```bash
# Only process personal namespaces (default: false)
export OKTETO_ONLY_PERSONAL_NAMESPACES=true

# Delete all unmounted PVCs instead of just dev volumes (default: true)
export OKTETO_DEV_VOLUMES=false
```

**Environment Variables:**
- `OKTETO_ONLY_PERSONAL_NAMESPACES`: When set to `true`, only processes namespaces with the label `dev.okteto.com/default-namespace=true` (personal namespaces). Default: `false`
- `OKTETO_DEV_VOLUMES`: When set to `true` (default), only deletes PVCs with label `dev.okteto.com=true`. When set to `false`, deletes ALL unmounted PVCs. Default: `true`

- Run the following command to create the cronjob:

**Basic deployment (default behavior - only dev volumes):**
```bash
okteto deploy -n ${NAMESPACE} --var OKTETO_ADMIN_TOKEN=${OKTETO_ADMIN_TOKEN} --var JOB_SCHEDULE=${JOB_SCHEDULE}
```

**Deploy with personal namespace filter:**
```bash
okteto deploy -n ${NAMESPACE} --var OKTETO_ADMIN_TOKEN=${OKTETO_ADMIN_TOKEN} --var JOB_SCHEDULE=${JOB_SCHEDULE} --var OKTETO_ONLY_PERSONAL_NAMESPACES=true
```

**Deploy to delete all unmounted PVCs (not just dev volumes):**
```bash
okteto deploy -n ${NAMESPACE} --var OKTETO_ADMIN_TOKEN=${OKTETO_ADMIN_TOKEN} --var JOB_SCHEDULE=${JOB_SCHEDULE} --var OKTETO_DEV_VOLUMES=false
```

**Deploy with both filters (personal namespaces + all PVCs):**
```bash
okteto deploy -n ${NAMESPACE} --var OKTETO_ADMIN_TOKEN=${OKTETO_ADMIN_TOKEN} --var JOB_SCHEDULE=${JOB_SCHEDULE} --var OKTETO_ONLY_PERSONAL_NAMESPACES=true --var OKTETO_DEV_VOLUMES=false
```

## Force the execution of the job

To force the execution of the job, run the following commands:

```bash
okteto kubeconfig
kubectl -n ${NAMESPACE} create job --from=cronjob/delete-dev-volumes delete-dev-volumes-$(date +%s)
```