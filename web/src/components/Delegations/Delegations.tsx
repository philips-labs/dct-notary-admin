import { FC, useEffect, useState, useCallback, useContext } from 'react';
import axios from 'axios';
import { useParams } from 'react-router-dom';
import { DelegationContext } from './DelegationContext';
import { RegisterDelegationKey } from './RegisterDelegationKey';
import { Delegation, DelegationListData } from '../../models';
import { TrashButton } from '..';
import { ApplicationContext } from '../Application';

const byRole = (a: Delegation, b: Delegation): number =>
  a.role < b.role ? -1 : a.role > b.role ? 1 : 0;

interface DelegationParams {
  targetId: string;
}

export const Delegations: FC = () => {
  const { targetId } = useParams<DelegationParams>();
  const { displayError, displayInfo } = useContext(ApplicationContext);
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
      setData((prevState) => ({
        ...prevState,
        delegations: [],
      }));
    }
  };

  const remove = async (delegation: Delegation) => {
    try {
      const response = await axios.delete(
        `/api/targets/${targetId}/delegations/${delegation.id.substr(0, 7)}`,
        {
          data: {
            delegationName: delegation.role,
          },
        },
      );
      displayInfo(
        `Removed delegation key with ID "${response.data.id}" for role "${response.data.role}"`,
        true,
      );
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
      <div className="mb-5 p-5 flex-none shadow-lg">
        <RegisterDelegationKey targetId={targetId} />
      </div>
      <ul>
        {data.delegations.map((item) => (
          <li className="flex flex-row justify-between px-6 py-3 align-middle border-t border-b border-gray-300 hover:bg-gray-50">
            <div>
              <span className="font-bold align-middle">{item.role}</span>
              <span className="ml-2 text-xs align-top italic">({item.id.substr(0, 7)})</span>
            </div>
            <TrashButton action={() => remove(item)} />
          </li>
        ))}
      </ul>
    </DelegationContext.Provider>
  ) : null;
};
