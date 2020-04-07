import React, { FC, useState, useEffect } from 'react';
import axios from 'axios';
import { RouteComponentProps } from 'react-router-dom';
import { Delegation, DelegationListData } from '../../models';

type TParams = { targetId: string };

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
        const delegations = [...delegationsResult.data];
        setData((prevState) => ({ ...prevState, delegations }));
      } catch (e) {
        setData((prevState) => ({ ...prevState, delegations: [] }));
      }
    };

    fetchData();
  }, [targetId]);

  return (
    <ul className="list-view">
      {data.delegations.map((item) => (
        <li key={item.id.substr(7)}>{item.role}</li>
      ))}
    </ul>
  );
};
