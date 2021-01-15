import React, { FC, cloneElement, MouseEvent } from 'react';
import { matchPath, useHistory, useLocation, useRouteMatch } from 'react-router';
import { Box, Text, Button, ButtonProps } from 'grommet';

export interface IconButton {
  label: string;
  icon: any;
}

export interface RoutedButtonProps {
  path: string;
}

export const NavButton: FC<RoutedButtonProps & ButtonProps & IconButton> = ({
  active,
  path,
  label,
  icon,
  ...rest
}) => {
  const match = useRouteMatch(path);
  const location = useLocation();
  const history = useHistory();

  const onClick = (event: MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    history.push(path);
  };

  const pathMatch = matchPath(location.pathname, { exact: match?.isExact, path });

  return (
    <Button
      active={active && !!pathMatch}
      {...rest}
      hoverIndicator={{ color: 'accent-1' }}
      plain
      onClick={onClick}
    >
      {({ hover }: { hover: boolean }) => (
        <Box pad={{ vertical: 'small' }} gap="xsmall" align="center" justify="center">
          {cloneElement(icon, { color: hover ? 'black' : 'white' })}
          <Text size="xsmall">{label}</Text>
        </Box>
      )}
    </Button>
  );
};
