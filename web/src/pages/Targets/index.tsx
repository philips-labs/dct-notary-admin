import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { Link, Route } from 'react-router-dom';
import { TargetListData, Target } from '../../models';
import { DelegationList } from '../../components/DelegationList';

export const TargetsPage: React.FC = () => {
  const [data, setData] = useState<TargetListData>({ targets: [] });

  useEffect(() => {
    const fetchData = async () => {
      const result = await axios.get<Target[]>('/api/targets');
      setData((prevState) => ({ ...prevState, targets: result.data }));
    };

    fetchData();
  }, []);

  return (
    <>
      <h2>Targets</h2>
      <ul className="list-view">
        {data.targets.map((item) => (
          <li key={item.id.substr(7)}>
            <Link to={`/targets/${item.id.substr(0, 7)}`}>{item.gun}</Link>
          </li>
        ))}
      </ul>
      <h3>Delegations</h3>
      <Route path="/targets/:targetId" component={DelegationList} />
    </>
  );
};
