import React, { FC, useEffect, useState, useContext } from 'react';
import axios from 'axios';
import { Route, useHistory } from 'react-router-dom';
import { Box, List } from 'grommet';
import { TargetListData, Target } from '../../models';
import { CreateTarget } from './CreateTarget';
import { TargetContext } from './TargetContext';
import { TrashButton } from '..';
import { ApplicationContext } from '../Application';

const byGun = (a: Target, b: Target): number => (a.gun < b.gun ? -1 : a.gun > b.gun ? 1 : 0);

export const Targets: FC = () => {
  const history = useHistory();
  const { displayError } = useContext(ApplicationContext);
  const [data, setData] = useState<TargetListData>({ targets: [] });
  const [selected, setSelected] = useState<number | undefined>();

  const fetchData = async () => {
    const result = await axios.get<Target[]>('/api/targets');
    const targets = [...result.data].sort(byGun);
    setData((prevState) => ({ ...prevState, targets }));
  };

  const remove = async (targetId: string) => {
    try {
      await axios.delete(`/api/targets/${targetId}`);
      fetchData();
    } catch (e) {
      displayError(`${e.message}: ${e.response.data}`, true);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <TargetContext.Provider value={{ refresh: fetchData }}>
      <Box margin={{ bottom: 'medium' }} elevation="medium" pad="medium" flex={false}>
        <Route path="/targets">
          <CreateTarget />
        </Route>
      </Box>
      <Box>
        {data.targets.length !== 0 ? (
          <List
            primaryKey="gun"
            secondaryKey={(item) => item.remove}
            itemProps={
              typeof selected !== 'undefined' && selected >= 0
                ? { [selected]: { background: 'accent-1' } }
                : undefined
            }
            data={data.targets.map((item) => ({
              ...item,
              remove: <TrashButton action={() => remove(item.id.substr(0, 7))} />,
            }))}
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
    </TargetContext.Provider>
  );
};
