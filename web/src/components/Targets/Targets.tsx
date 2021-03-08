import { FC, useEffect, useState, useContext } from 'react';
import axios from 'axios';
import { Route, useHistory } from 'react-router-dom';
import { TargetListData, Target } from '../../models';
import { CreateTarget } from './CreateTarget';
import { TargetContext } from './TargetContext';
import { TrashButton } from '..';
import { ApplicationContext } from '../Application';
import cn from 'classnames';

const byGun = (a: Target, b: Target): number => (a.gun < b.gun ? -1 : a.gun > b.gun ? 1 : 0);

export const Targets: FC = () => {
  const history = useHistory();
  const { displayError } = useContext(ApplicationContext);
  const [data, setData] = useState<TargetListData>({ targets: [] });

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
      <div className="mb-5 p-5 flex-none shadow-lg">
        <Route path="/targets">
          <CreateTarget />
        </Route>
      </div>
      <div>
        {data.targets.length !== 0 ? (
          <>
            <ul>
              {data.targets.map((item, i) => (
                <li
                  key={i}
                  className={cn('flex flex-row justify-between px-6 py-3 align-middle', {
                    'border-gray-300 border-t border-b hover:bg-gray-50': !history.location.pathname.endsWith(
                      item.id.substr(0, 7),
                    ),
                    'bg-blue-200 border-blue-400 border-2': history.location.pathname.endsWith(
                      item.id.substr(0, 7),
                    ),
                  })}
                  onClick={() => {
                    history.push(`/targets/${item.id.substr(0, 7)}`);
                  }}
                >
                  <div className="font-bold align-middle">{item.gun}</div>
                  <TrashButton action={() => remove(item.id.substr(0, 7))} />
                </li>
              ))}
            </ul>
          </>
        ) : (
          'Loading...'
        )}
      </div>
    </TargetContext.Provider>
  );
};
