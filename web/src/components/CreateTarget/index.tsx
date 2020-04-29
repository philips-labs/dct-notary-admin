import React, { FC, useState, FormEvent } from 'react';
import { RouteComponentProps } from 'react-router-dom';
import { Box, Form, TextInput, Button, Text } from 'grommet';
import axios from 'axios';
import { FormFieldLabel } from '..';

type TParams = { targetId: string };

type CreateTarget = { gun: string; errorMessage: string };
const defaultFormValue = { gun: '', errorMessage: '' };

export const CreateTarget: FC<RouteComponentProps<TParams>> = () => {
  const [value, setValue] = useState<CreateTarget>(defaultFormValue);
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
        setValue(event as CreateTarget);
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
