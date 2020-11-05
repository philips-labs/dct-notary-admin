import React, { FC, useState, useContext, FormEvent } from 'react';
import { Box, Form, TextInput, Button, Text } from 'grommet';
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
      {value.errorMessage && (
        <Box pad={{ horizontal: 'small' }}>
          <Text color="status-error">{value.errorMessage}</Text>
        </Box>
      )}
      <Box direction="row" justify="end" margin={{ top: 'medium' }}>
        <Button type="submit" label="Submit" primary />
      </Box>
    </Form>
  );
};
