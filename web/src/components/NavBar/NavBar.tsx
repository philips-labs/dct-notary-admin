import { FC, ClassAttributes, HTMLAttributes } from 'react';
import { BoxProps, Sidebar, Nav, Box, Avatar } from 'grommet';
import { Home, Database } from 'grommet-icons';
import { NavButton } from './NavButton';

export const MainNavigation = () => (
  <Nav pad={{ vertical: 'small' }}>
    <NavButton path="/targets" icon={<Database />} label="Targets" />
  </Nav>
);

export const SidebarHeader = () => (
  <Nav gap="small">
    <NavButton path="/" icon={<Home />} label="Home" />
  </Nav>
);

export const SidebarFooter = () => (
  <Box pad="small" border={{ color: 'white', side: 'top' }}>
    <Avatar border={{ size: 'small', color: 'accent-2' }} background="white" flex={false}>
      MF
    </Avatar>
  </Box>
);

export const NavBar: FC<
  BoxProps & ClassAttributes<HTMLDivElement> & HTMLAttributes<HTMLDivElement>
> = () => {
  return (
    <Sidebar
      background="brand"
      pad={{ left: 'none', right: 'none' }}
      header={<SidebarHeader />}
      footer={<SidebarFooter />}
      gap="small"
    >
      <MainNavigation />
    </Sidebar>
  );
};
