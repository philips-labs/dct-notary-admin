import React, { FC, useEffect, useState, useCallback, useContext } from 'react';
import axios from 'axios';
import { useParams } from 'react-router-dom';
import { Box, List } from 'grommet';
import { DelegationContext } from './DelegationContext';
import { RegisterDelegationKey } from './RegisterDelegationKey';
import { Delegation, DelegationListData } from '../../models';
import { TrashButton } from '..';
import { ApplicationContext } from '../Application';

const byRole = (a: Delegation, b: Delegation): number =>
  a.role < b.role ? -1 : a.role > b.role ? 1 : 0;

export const Delegations: FC = () => {
  const { targetId } = useParams();
  const { displayError } = useContext(ApplicationContext);
  const [data, setData] = useState<DelegationListData>({
    delegations: [],
  });

  const fetchData = async () => {
    try {
      const delegationsResult = await axios.get<Delegation[]>(
        `/api/targets/${targetId}/delegations`,
      );
      const delegations = [...delegationsResult.data].sort(byRole);
      setData((prevState) => ({ ...prevState, delegations }));
    } catch (e) {
      setData((prevState) => ({ ...prevState, delegations: [] }));
    }
  };

  const remove = async (delegationId: string) => {
    try {
      await axios.delete(`/api/targets/${targetId}/delegations/${delegationId}`);
      fetchData();
    } catch (e) {
      displayError(`${e.message}: ${e.response.data}`, true);
    }
  };

  const fetchDataCallback = useCallback(fetchData, [targetId]);
  useEffect(() => {
    fetchDataCallback();
  }, [fetchDataCallback]);

  return targetId ? (
    <DelegationContext.Provider value={{ refresh: fetchData }}>
      <Box margin={{ bottom: 'medium' }} elevation="medium" pad="medium" flex={false}>
        <RegisterDelegationKey targetId={targetId} />
      </Box>
      <Box>
        <List
          primaryKey="role"
          secondaryKey={(item) => item.remove}
          data={data.delegations.map((item) => ({
            ...item,
            remove: <TrashButton action={() => remove(item.id.substr(0, 7))} />,
          }))}
        />
      </Box>
    </DelegationContext.Provider>
  ) : null;
};
