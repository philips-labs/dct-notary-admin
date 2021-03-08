import { FC, useState, useContext, FormEvent, ChangeEvent } from 'react';
import axios from 'axios';
import { TargetContext } from './TargetContext';
import { FormTextInput } from '../Form';

type CreateTargetState = { gun: string; errorMessage?: string };
const defaultFormValue = { gun: '', errorMessage: '' };

export const CreateTarget: FC = () => {
  const [value, setValue] = useState<CreateTargetState>(defaultFormValue);
  const { refresh } = useContext(TargetContext);
  const submitForm = async (event: FormEvent) => {
    event.preventDefault();
    try {
      const { errorMessage, ...requestBody } = value;
      await axios.post(`/api/targets`, JSON.stringify(requestBody), {
        headers: new Headers({
          'Content-Type': 'application/json',
          Accept: 'application/json',
        }),
      });
      setValue(defaultFormValue);
      refresh();
    } catch (e) {
      const response = e.response;
      const errorMessage = `${response.data.status} ${response.data.error}`;
      setValue({ ...value, errorMessage });
      console.log(value);
    }
  };

  return (
    <form onSubmit={submitForm} className="antialiased">
      <FormTextInput
        label="GUN"
        name="gun"
        placeholder="docker.io/philipssoftware/openjdk"
        required
        onChange={(event: ChangeEvent<HTMLInputElement>) => {
          setValue({ gun: event.target.value });
        }}
        className="mb-2"
      />
      {value.errorMessage && <p className="text-sm text-red-500 p-1">{value.errorMessage}</p>}
      <div className="flex flex-row-reverse">
        <button
          className="bg-blue-600 text-white p-2 px-5 hover:bg-blue-700 rounded-3xl font-semibold"
          type="submit"
        >
          Submit
        </button>
      </div>
    </form>
  );
};
