import React, { FC, useEffect, useState } from 'react';
import axios from 'axios';
import { useParams } from 'react-router-dom';
import { Box, List } from 'grommet';
import { DelegationContext } from './DelegationContext';
import { RegisterDelegationKey } from './RegisterDelegationKey';
import { Delegation, DelegationListData } from '../../models';

const byRole = (a: Delegation, b: Delegation): number =>
  a.role < b.role ? -1 : a.role > b.role ? 1 : 0;

export const Delegations: FC = () => {
  const { targetId } = useParams();
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

  useEffect(() => {
    fetchData();
  }, [targetId]);

  return targetId ? (
    <DelegationContext.Provider value={{ refresh: fetchData }}>
      <Box margin={{ bottom: 'medium' }} elevation="medium" pad="medium" flex={false}>
        <RegisterDelegationKey targetId={targetId} />
      </Box>
      <Box>
        <List
          primaryKey="role"
          secondaryKey={(item) => item.id.substr(0, 7)}
          data={data.delegations}
        />
      </Box>
    </DelegationContext.Provider>
  ) : (
    <></>
  );
};
