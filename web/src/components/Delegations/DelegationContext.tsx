import { createContext } from 'react';

export const DelegationContext = createContext({
  refresh: () => {},
});
