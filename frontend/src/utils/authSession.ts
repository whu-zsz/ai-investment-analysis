export function clearAuthSession() {
  localStorage.removeItem('token');
  localStorage.removeItem('userInfo');
}
