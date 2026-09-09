import { $useAuth, $lib } from "@typespec/http";

export function $usePermission(context, entity, auth) {
    $useAuth(context, entity, auth)
    const { options } = context.program.stateMap($lib.stateKeys.authentication).get(entity)
    const schemes = options.map(option => option.schemes).flat()
    const permissions = schemes.filter((scheme, index, self) => self.findIndex(t => t.id === scheme.id) === index)
    context.program.stateMap($lib.stateKeys.authentication).set(entity, { options: [{ schemes: permissions }] })
}
