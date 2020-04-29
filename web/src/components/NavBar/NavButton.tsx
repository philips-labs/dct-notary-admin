import React, { FC } from 'react';
import { Box, Text, RoutedButton } from 'grommet';

export interface NavButtonProps {
  path: string;
  icon: any;
  label: string;
}

export const NavButton: FC<NavButtonProps> = ({ path, icon, label }) => {
  const tooltipColor = { color: 'accent-1', opacity: 0.9 };
  return (
    <RoutedButton path={path} hoverIndicator={tooltipColor}>
      <Box pad={{ vertical: 'small' }} gap="xsmall" align="center" justify="center">
        {icon}
        <Text size="xsmall">{label}</Text>
      </Box>
    </RoutedButton>
  );
};
