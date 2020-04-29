import React, { FC, useState, useEffect } from 'react';
import axios from 'axios';
import { RouteComponentProps } from 'react-router-dom';
import { Delegation, DelegationListData } from '../../models';
import { List } from 'grommet';

type TParams = { targetId: string };

const byRole = (a: Delegation, b: Delegation): number =>
  a.role < b.role ? -1 : a.role > b.role ? 1 : 0;

export const DelegationList: FC<RouteComponentProps<TParams>> = ({ match }) => {
  const { targetId } = match.params;
  const [data, setData] = useState<DelegationListData>({
    delegations: [],
  });

  useEffect(() => {
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

    fetchData();
  }, [targetId]);

  return (
    <List primaryKey="role" secondaryKey={(item) => item.id.substr(0, 7)} data={data.delegations} />
  );
};
