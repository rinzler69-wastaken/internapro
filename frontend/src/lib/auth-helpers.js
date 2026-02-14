export function needsProfileCompletion(user) {
  if (!user) return true;

  const role = user.role || user?.user?.role || '';
  const internIdFromUser = user.intern_id || user.InternID || user.internId;
  if (role === 'new_user') {
    const intern = user.intern || user?.profile?.intern || user?.user?.intern;
    // If an intern profile already exists and is not pending, treat as complete to avoid loops with stale tokens.
    if (
      (internIdFromUser && (!intern || intern?.status?.toLowerCase() !== 'pending')) ||
      (intern && intern.id && intern.status && intern.status.toLowerCase() !== 'pending')
    ) {
      return false;
    }
    return true;
  }

  const intern = user.intern || user?.profile?.intern || user?.user?.intern;
  if (role === 'intern' && (!intern || !intern.id) && !internIdFromUser) return true;

  return false;
}
