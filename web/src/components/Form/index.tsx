import React, { FC, ReactNode, FormEvent, useState, createContext } from 'react';
import axios from 'axios';
import { FieldProps } from './field';
export interface IFormContext {
  setValues: (values: FormValues) => void;
  errors: FormErrors;
  validate: (fieldName: string) => void;
}

export const FormContext = createContext<IFormContext | undefined>(undefined);

export interface Fields {
  [key: string]: FieldProps;
}

export interface FormProps {
  action: string;
  fields: Fields;
  render: () => ReactNode;
}

export interface FormValues {
  [key: string]: any;
}

export interface FormErrors {
  [key: string]: string;
}

export interface FormState {
  values: FormValues;
  errors: FormErrors;
  submitSuccess?: boolean;
}

export const Form: FC<FormProps> = ({ action, fields, render }) => {
  const [state, setState] = useState<FormState>({
    values: {},
    errors: {},
  });
  const { submitSuccess, errors } = state;

  const setValues = (values: FormValues) => {
    setState({ ...state, values: { ...state.values, ...values } });
  };

  const validateForm = (): boolean => {
    const errors: FormErrors = {};
    Object.keys(fields).forEach((fieldName: string) => {
      errors[fieldName] = validate(fieldName);
    });
    setState({ ...state, errors });
    return !hasErrors(errors);
  };

  const validate = (fieldName: string): string => {
    let newError: string = '';

    if (fields[fieldName] && fields[fieldName].validator) {
      newError = fields[fieldName].validator!.rule(
        state.values,
        fieldName,
        fields[fieldName].validator!.args,
      );
    }
    state.errors[fieldName] = newError;
    setState({ ...state, errors: { ...state.errors, [fieldName]: newError } });
    return newError;
  };

  const context: IFormContext = {
    ...state,
    setValues,
    validate,
  };

  const hasErrors = (error: FormErrors): boolean => {
    let haveError: boolean = false;
    Object.keys(errors).forEach((key: string) => {
      if (errors[key].length > 0) {
        haveError = true;
      }
    });
    return haveError;
  };

  const submitForm = async (): Promise<boolean> => {
    try {
      const response = await axios.post(action, JSON.stringify(state.values), {
        headers: new Headers({
          'Content-Type': 'application/json',
          Accept: 'application/json',
        }),
      });

      return response && response.statusText === 'OK';
    } catch (ex) {
      return false;
    }
  };

  const handleSubmit = async (e: FormEvent<HTMLFormElement>): Promise<void> => {
    e.preventDefault();
    if (validateForm()) {
      const submitSuccess: boolean = await submitForm();
      setState({ ...state, submitSuccess });
    }
  };

  return (
    <FormContext.Provider value={context}>
      <div className="form-container">
        <form onSubmit={handleSubmit} noValidate={true}>
          {render()}
          <button type="submit" disabled={hasErrors(errors)}>
            Submit
          </button>
          {submitSuccess && (
            <div className="alert alert-info" role="alert">
              The form was successfully submitted!
            </div>
          )}
          {submitSuccess === false && !hasErrors(errors) && (
            <div className="alert alert-danger" role="alert">
              Sorry, an unexpected error has occurred
            </div>
          )}
          {submitSuccess === false && hasErrors(errors) && (
            <div className="alert alert-danger" role="alert">
              Sorry, the form is invalid. Please review, adjust and try again
            </div>
          )}
        </form>
      </div>
    </FormContext.Provider>
  );
};

export * from './field';
