import { ChangeEventHandler } from 'react';
import cn from 'classnames';

const uniqueId = ((): ((prefix: string) => string) => {
  let counter = 0;
  return (prefix: string): string => `${prefix}${++counter}`;
})();

interface FormElementProps {
  label: string;
  name: string;
  placeholder?: string;
  value?: string;
  help?: string;
  required: boolean;
  className?: string;
}
interface FormInputProps {
  onChange?: ChangeEventHandler<HTMLInputElement>;
}

interface FormTextAreaProps {
  onChange?: ChangeEventHandler<HTMLTextAreaElement>;
}

export function FormTextInput({
  label,
  name,
  placeholder,
  required,
  value,
  help,
  className,
  onChange,
}: FormElementProps & FormInputProps) {
  const id = uniqueId(`input-${name?.replaceAll(' ', '').toLowerCase()}`);

  return (
    <label htmlFor={id} className={cn('block text-gray-800', className)}>
      <span className="mr-2">{label}</span>
      {required && <span className="text-red-500 ml-1">*</span>}
      <span className="block text-gray-500 text-sm">{help}</span>
      <input
        id={id}
        type="text"
        name={name}
        className="block border-0 border-b-2 border-gray-200 focus:ring-0 focus:border-blue-400 px-0.5 w-full"
        placeholder={placeholder}
        value={value}
        onChange={onChange}
      />
    </label>
  );
}

export function FormTextArea({
  label,
  name,
  placeholder,
  required,
  value,
  help,
  className,
  onChange,
}: FormElementProps & FormTextAreaProps) {
  const id = uniqueId(`input-${name?.replaceAll(' ', '').toLowerCase()}`);

  return (
    <label htmlFor={id} className={cn('block text-gray-800', className)}>
      <span className="mr-2">{label}</span>
      {required && <span className="text-red-500 ml-1">*</span>}
      <span className="block text-gray-500 text-sm">{help}</span>
      <textarea
        id={id}
        name={name}
        className="block border-0 border-b-2 border-gray-200 focus:ring-0 focus:border-blue-400 px-0.5 w-full h-36"
        placeholder={placeholder}
        value={value}
        onChange={onChange}
      />
    </label>
  );
}
