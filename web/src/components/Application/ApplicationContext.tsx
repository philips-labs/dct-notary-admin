import { createContext } from 'react';

export const ApplicationContext = createContext({
  displayError: (message: string, autoHide: boolean) => {},
});
