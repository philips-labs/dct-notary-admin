import { FC } from 'react';
import { Route } from 'react-router-dom';
import { Delegations, Targets } from '../../components';

export const TargetsPage: FC = () => {
  return (
    <div className="overflow-auto grid grid-cols-1 lg:grid-cols-2 gap-10">
      <div>
        <h2 className="text-4xl tracking-tight font-bold">Targets</h2>
        <Targets />
      </div>
      <div>
        <h2 className="text-4xl tracking-tight font-bold">Delegations</h2>
        <Route path="/targets/:targetId">
          <Delegations />
        </Route>
      </div>
    </div>
  );
};
