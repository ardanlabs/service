'use client'

import * as React from 'react'
import Typography from '@mui/material/Typography'
import Box from '@mui/system/Box'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import { Modal } from '@/components/Modal/Modal'
import Autocomplete from '@mui/material/Autocomplete'
import PasswordTextField from '@/components/PasswordTextField/PasswordTextField'

export default function AddUser() {
  const [open, setOpen] = React.useState(false)
  const handleOpen = () => setOpen(true)
  const handleClose = () => setOpen(false)

  interface formDataInterface {
    name: string
    email: string
    roles: string[]
    department: string
    password: string
    passwordConfirm: string
  }

  interface formDataErrorInterface {
    name: { value: boolean; message: string }
    email: { value: boolean; message: string }
    roles: { value: boolean; message: string }
    department: { value: boolean; message: string }
    password: { value: boolean; message: string }
    passwordConfirm: { value: boolean; message: string }
  }

  const [formData, setFormData] = React.useState<formDataInterface>({
    name: '',
    email: '',
    roles: [],
    department: '',
    password: '',
    passwordConfirm: '',
  })

  const [errors, setErrors] = React.useState<formDataErrorInterface>({
    name: { value: false, message: '' },
    email: { value: false, message: '' },
    roles: { value: false, message: '' },
    department: { value: false, message: '' },
    password: { value: false, message: '' },
    passwordConfirm: { value: false, message: '' },
  })

  const emailRule = (): boolean => {
    if (!formData.email.length) {
      setErrors((prevFormErrors) => ({
        ...prevFormErrors,
        email: { value: true, message: 'This field is required' },
      }))
      return true
    }
    if (formData.email.length <= 6 || formData.email.length >= 128) {
      setErrors((prevFormErrors) => ({
        ...prevFormErrors,
        email: { value: true, message: 'Value out of range' },
      }))
      return true
    }
    const emailRegExp =
      /^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/

    if (!emailRegExp.test(formData.email)) {
      setErrors((prevFormErrors) => ({
        ...prevFormErrors,
        email: { value: true, message: 'Email Invalid' },
      }))
      return true
    }
    setErrors((prevFormErrors) => ({
      ...prevFormErrors,
      email: { value: false, message: '' },
    }))
    setFormData((prevFormData) => ({ ...prevFormData, email: formData.email }))
    return false
  }

  const availableRoles = ['USER', 'ADMIN']

  const handleRolesChange = (
    _event: React.SyntheticEvent,
    newValue: string[],
  ) => {
    setFormData((prevFormData) => ({ ...prevFormData, roles: newValue }))
  }

  const validate = () => {
    let temp: formDataErrorInterface = { ...errors }

    console.log(formData)

    if ('name' in formData)
      temp.name = formData.name
        ? { value: false, message: '' }
        : { value: true, message: 'This field is required.' }

    if ('roles' in formData)
      temp.roles = formData.roles.length
        ? { value: false, message: '' }
        : { value: true, message: 'This field is required.' }

    if ('password' in formData)
      temp.password = formData.password
        ? { value: false, message: '' }
        : { value: true, message: 'This field is required.' }

    if ('passwordConfirm' in formData)
      temp.passwordConfirm = formData.passwordConfirm
        ? { value: false, message: '' }
        : { value: true, message: 'This field is required.' }

    console.log(temp)

    setErrors(temp)
  }

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = event.target
    setFormData((prevFormData) => ({ ...prevFormData, [name]: value }))
  }

  const handleSubmit = async () => {
    validate()
    emailRule()
  }
  return (
    <Modal
      buttonText="Add User"
      handleOpen={handleOpen}
      handleClose={handleClose}
      open={open}
    >
      <Box
        sx={{
          display: 'flex',
          alignContent: 'center',
          justifyContent: 'center',
          flexDirection: 'column',
        }}
      >
        <Typography id="modal-modal-title" variant="h6" component="h2">
          Add User
        </Typography>
        <form>
          <Box
            id="modal-modal-content"
            sx={{
              mt: 2,
              display: 'flex',
              alignContent: 'center',
              justifyContent: 'start',
              flexDirection: 'column',
            }}
          >
            <TextField
              required
              id="name"
              label="Name"
              name="name"
              variant="filled"
              error={errors.name.value}
              helperText={errors.name.message}
              sx={{
                backgroundColor: '#FFFFFF',
                borderRadius: '4px',
                m: 1,
              }}
              onChange={handleChange}
            />
            <TextField
              required
              id="email"
              label="Email"
              name="email"
              variant="filled"
              error={errors.email.value}
              helperText={errors.email.message}
              sx={{
                backgroundColor: '#FFFFFF',
                borderRadius: '4px',
                m: 1,
              }}
              onChange={emailRule}
            />
            <Autocomplete
              disablePortal
              multiple
              value={formData.roles}
              sx={{
                backgroundColor: '#FFFFFF',
                borderRadius: '4px',
                m: 1,
              }}
              options={availableRoles}
              renderInput={(params) => (
                <TextField
                  {...params}
                  id="roles"
                  error={errors.roles.value}
                  helperText={errors.roles.message}
                  variant="filled"
                  label="Role*"
                />
              )}
              onChange={handleRolesChange}
            />
            <TextField
              id="department"
              label="Department"
              name="department"
              variant="filled"
              sx={{
                backgroundColor: '#FFFFFF',
                borderRadius: '4px',
                m: 1,
              }}
              onChange={handleChange}
            />
            <PasswordTextField
              label="Password*"
              name="passwordConfirm"
              error={errors.password.value}
              helperText={errors.password.message}
              styles={{ m: 1 }}
              handleOnChange={handleChange}
            />
            <PasswordTextField
              label="Confirm Password*"
              name="passwordConfirm"
              error={errors.passwordConfirm.value}
              helperText={errors.passwordConfirm.message}
              styles={{ m: 1 }}
              handleOnChange={handleChange}
            />
          </Box>
        </form>
        <Button
          type="submit"
          variant="contained"
          color="primary"
          sx={{ my: 2, alignSelf: 'end' }}
          onClick={handleSubmit}
        >
          Add
        </Button>
      </Box>
    </Modal>
  )
}
