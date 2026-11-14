// Test-only module substitution. The transaction repair component itself remains unmodified.
// This loader replaces the authenticated context boundary with a fixed test principal;
// real React, the real transaction repair component, and the real API/idempotency helpers execute.
export async function resolve(specifier, context, nextResolve) {
  if (specifier === "@/features/auth/components/AuthShell" ||
      /(?:^|\/)features\/auth\/components\/AuthShell(?:\.tsx)?$/.test(specifier)) {
    const body = 'export function useAuth(){return {principalId:"00000000-0000-4000-8000-000000000001"}}';
    return {url: `data:text/javascript,${encodeURIComponent(body)}`, shortCircuit: true};
  }
  return nextResolve(specifier, context);
}
