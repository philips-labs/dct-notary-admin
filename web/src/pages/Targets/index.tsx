import { FC } from 'react';
import { Route } from 'react-router-dom';
import { Heading, Box, Grid } from 'grommet';
import { Delegations, Targets } from '../../components';

export const TargetsPage: FC = () => {
  return (
    <Box overflow="auto">
      <Grid
        fill
        rows={['auto']}
        columns={['flex', 'flex']}
        areas={[
          { name: 'targets', start: [0, 0], end: [0, 0] },
          { name: 'delegations', start: [1, 0], end: [1, 0] },
        ]}
      >
        <Box gridArea="targets" pad="medium" overflow="auto" fill>
          <Heading level={2}>Targets</Heading>
          <Targets />
        </Box>
        <Box gridArea="delegations" pad="medium" overflow="auto">
          <Heading level={2}>Delegations</Heading>
          <Route path="/targets/:targetId">
            <Delegations />
          </Route>
        </Box>
      </Grid>
    </Box>
  );
};
