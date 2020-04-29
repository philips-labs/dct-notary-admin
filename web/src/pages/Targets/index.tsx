import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Route, useHistory } from 'react-router-dom';
import { Heading, Box, List, Grid } from 'grommet';
import { TargetListData, Target } from '../../models';
import { CreateTarget, RegisterDelegationKey, DelegationList } from '../../components';

const byGun = (a: Target, b: Target): number => (a.gun < b.gun ? -1 : a.gun > b.gun ? 1 : 0);

export const TargetsPage: React.FC = () => {
  const history = useHistory();
  const [data, setData] = useState<TargetListData>({ targets: [] });
  const [selected, setSelected] = useState<number | undefined>();

  useEffect(() => {
    const fetchData = async () => {
      const result = await axios.get<Target[]>('/api/targets');
      const targets = [...result.data].sort(byGun);
      setData((prevState) => ({ ...prevState, targets }));
    };

    fetchData();
  }, []);

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
          <Box margin={{ bottom: 'medium' }} elevation="medium" pad="medium" flex={false}>
            <Route path="/targets" component={CreateTarget} />
          </Box>
          <Box>
            {data.targets.length !== 0 ? (
              <List
                primaryKey="gun"
                secondaryKey={(item) => item.id.substr(0, 7)}
                itemProps={
                  typeof selected !== 'undefined' && selected >= 0
                    ? { [selected]: { background: 'accent-1' } }
                    : undefined
                }
                data={data.targets}
                onClickItem={(event: { item?: {}; index?: number }) => {
                  setSelected(selected === event.index ? undefined : event.index);
                  const item: { id: string } | undefined = event.item as { id: string };
                  if (item) {
                    history.push(`/targets/${item.id.substr(0, 7)}`);
                  }
                }}
              />
            ) : (
              'Loading...'
            )}
          </Box>
        </Box>
        <Box gridArea="delegations" pad="medium" overflow="auto">
          <Heading level={2}>Delegations</Heading>
          <Box margin={{ bottom: 'medium' }} elevation="medium" pad="medium" flex={false}>
            <Route path="/targets/:targetId" component={RegisterDelegationKey} />
          </Box>
          <Box>
            <Route path="/targets/:targetId" component={DelegationList} />
          </Box>
        </Box>
      </Grid>
    </Box>
  );
};
