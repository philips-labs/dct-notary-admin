import React, { FC, cloneElement } from 'react';
import { Box, Text, RoutedButton, RoutedButtonProps, ButtonProps } from 'grommet';

export interface IconButton {
  label: string;
  icon: any;
}

export const NavButton: FC<RoutedButtonProps & ButtonProps & IconButton> = ({
  path,
  label,
  icon,
  ...rest
}) => {
  return (
    <RoutedButton path={path} {...rest} hoverIndicator={{ color: 'accent-1' }} plain>
      {({ hover }: { hover: boolean }) => (
        <Box pad={{ vertical: 'small' }} gap="xsmall" align="center" justify="center">
          {cloneElement(icon, { color: hover ? 'black' : 'white' })}
          <Text size="xsmall">{label}</Text>
        </Box>
      )}
    </RoutedButton>
  );
};
