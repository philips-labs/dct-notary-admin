import React, { FC, ClassAttributes, HTMLAttributes, useContext } from 'react';
import { ResponsiveContext, BoxProps, Sidebar, Nav } from 'grommet';
import { Home, Database, User } from 'grommet-icons';
import { NavButton } from './NavButton';

export const NavBar: FC<
  BoxProps & ClassAttributes<HTMLDivElement> & HTMLAttributes<HTMLDivElement>
> = () => {
  const size = useContext(ResponsiveContext);

  return (
    <Sidebar
      background="brand"
      header={<NavButton path="/" icon={<Home />} label="Home" />}
      footer={<NavButton path="/" icon={<User />} label="User" />}
    >
      <Nav align="center" pad={{ vertical: 'small' }} gap={size === 'small' ? 'medium' : 'small'}>
        <NavButton path="/targets" icon={<Database />} label="Targets" />
      </Nav>
    </Sidebar>
  );
};
