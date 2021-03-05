import { FC, useState, useContext, FormEvent } from 'react';
import { Form, TextInput } from 'grommet';
import axios from 'axios';
import { FormFieldLabel } from '..';
import { TargetContext } from './TargetContext';

type CreateTargetState = { gun: string; errorMessage: string };
const defaultFormValue = { gun: '', errorMessage: '' };

export const CreateTarget: FC = () => {
  const [value, setValue] = useState<CreateTargetState>(defaultFormValue);
  const { refresh } = useContext(TargetContext);
  const submitForm = async (event: FormEvent) => {
    event.preventDefault();
    try {
      await axios.post(`/api/targets`, JSON.stringify(value), {
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
    <Form
      value={value}
      onChange={(event: any) => {
        setValue(event as CreateTargetState);
      }}
      onSubmit={submitForm}
      validate="blur"
    >
      <FormFieldLabel label="GUN" name="gun" required>
        <TextInput name="gun" placeholder="docker.io/philipssoftware/openjdk" required />
      </FormFieldLabel>
      {value.errorMessage && <p className="text-sm text-red-500 p-1">{value.errorMessage}</p>}
      <div className="flex flex-row-reverse">
        <button
          className="bg-blue-600 text-white p-2 px-5 hover:bg-blue-700 rounded-3xl font-semibold"
          type="submit"
        >
          Submit
        </button>
      </div>
    </Form>
  );
};
