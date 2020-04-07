import React, { useState, useEffect } from 'react';
import axios from 'axios';
interface Target {
  id: string;
  gun: string;
  role: string;
}

interface ApiData {
  targets: Target[];
}

export const TargetsPage: React.FC = () => {
  const [data, setData] = useState<ApiData>({ targets: [] });

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
      <ul className="Targets">
        {data.targets.map((item) => (
          <li key={item.id}>
            <a href="./">{item.gun}</a>
          </li>
        ))}
      </ul>
    </>
  );
};
