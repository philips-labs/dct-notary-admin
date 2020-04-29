import { grommet } from 'grommet/themes';
import { deepMerge } from 'grommet/utils';

export const customTheme = deepMerge(grommet, {
  global: {
    breakpoints: {
      xsmall: {
        value: 400,
      },
    },
    colors: {
      brand: '#035ed8',
    },
  },
  paragraph: {
    extend: () => `font-weight: 300; margin-top: 0;`,
    xxlarge: {
      size: '28px',
    },
  },
});
