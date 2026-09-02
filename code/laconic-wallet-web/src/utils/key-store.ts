const setInternetCredentials = (name:string, username:string, password:string) => {
  localStorage.setItem(name, password);
};

const getInternetCredentials =  (name:string) : string | null => {
  return localStorage.getItem(name);
};

const resetInternetCredentials = (name:string) => {
  localStorage.removeItem(name);
};

export {
  setInternetCredentials,
  getInternetCredentials,
  resetInternetCredentials
}
