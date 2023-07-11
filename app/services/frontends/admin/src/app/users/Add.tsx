'use client'

import * as React from 'react'
import Typography from '@mui/material/Typography'
import Box from '@mui/system/Box'
import TextField from '@mui/material/TextField'
import Button from '@mui/material/Button'
import { Modal } from '@/components/Modal/Modal'
import Autocomplete from '@mui/material/Autocomplete'
import PasswordTextField from '@/components/PasswordTextField/PasswordTextField'
import Chip from '@mui/material/Chip'
import ApiError from '@/components/ApiError/ApiError'

interface AddUserProps {
  setNeedsUpdate?: React.Dispatch<React.SetStateAction<boolean>>
}

export default function AddUser(props: AddUserProps) {
  const { setNeedsUpdate } = props
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

  type dataError = { value: boolean; message: string }

  interface formDataErrorInterface {
    name: dataError
    email: dataError
    roles: dataError
    department: dataError
    password: dataError
    passwordConfirm: dataError
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

  const [fetchError, setFetchError] = React.useState('')

  function emailRule(): dataError {
    if (!formData.email.length) {
      return { value: true, message: 'This field is required' }
    }
    if (formData.email.length <= 6 || formData.email.length >= 128) {
      return { value: true, message: 'Value out of range' }
    }
    const emailRegExp =
      /^[a-zA-Z0-9.!#$%&’*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/

    if (!emailRegExp.test(formData.email)) {
      return { value: true, message: 'Email Invalid' }
    }
    return { value: false, message: '' }
  }

  function validate(data: Partial<formDataInterface>): boolean {
    const temp: formDataErrorInterface = { ...errors }
    let isValid = true

    let field: keyof Partial<formDataInterface> // Gets the keys of the form data.
    for (field in data) {
      const value = data[field] // Extracts value.
      if (field === 'email') {
        // If email, runs email rule.
        isValid = !emailRule().value
        continue
      }

      if (value && !value.length) {
        // Checks if value is required.
        isValid = false
        temp[field] = { value: true, message: 'This field is required.' }
        continue
      }

      if (field === 'passwordConfirm' || field === 'password') {
        // Checks if the two password fields match.
        if (formData.passwordConfirm !== formData.password) {
          isValid = false
          temp.password = { value: true, message: 'Passwords need to match' }
          temp.passwordConfirm = {
            value: true,
            message: 'Passwords need to match',
          }
          continue
        }
        if (formData.passwordConfirm === formData.password) {
          isValid = true
          temp.password = { value: false, message: '' }
          temp.passwordConfirm = { value: false, message: '' }
          continue
        }
      }

      // If value is clean, we set the error to it's zero value.
      temp[field] = { value: false, message: '' }
    }

    setErrors(temp)
    return isValid
  }

  const availableRoles = ['USER', 'ADMIN']

  function handleRolesChange(_event: React.SyntheticEvent, newValue: string[]) {
    setFormData((prevFormData) => ({ ...prevFormData, roles: newValue }))
  }

  function handleChange(event: React.ChangeEvent<HTMLInputElement>) {
    const { name, value } = event.target
    if (name === 'email') emailRule()
    setFormData((prevFormData) => ({ ...prevFormData, [name]: value }))
  }

  async function handleSubmit() {
    // Before we submit we validate the form
    const isValid = validate(formData)

    if (isValid) {
      try {
        const userPost = await fetch(
          `${process.env.NEXT_PUBLIC_BASE_API_URL}/users`,
          {
            method: 'POST',
            body: JSON.stringify(formData),
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${process.env.NEXT_PUBLIC_TOKEN}`,
            },
          },
        )
        if (userPost.ok) {
          setFormData({
            name: '',
            email: '',
            roles: [],
            department: '',
            password: '',
            passwordConfirm: '',
          })
          setOpen(false)
          if (setNeedsUpdate) {
            setNeedsUpdate(true)
          }
          return
        }

        const error: { error: string } = await userPost.json()

        setFetchError(error.error)
        setFormData((prevFormData) => ({ ...prevFormData, roles: [] }))
      } catch (error) {
        console.log(error)
      }
    }
  }
  return (
    <Modal
      buttonText="Add User"
      handleOpen={handleOpen}
      handleClose={handleClose}
      open={open}
    >
      {fetchError ? (
        <ApiError message={fetchError} clearError={() => setFetchError('')} />
      ) : (
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
                onChange={handleChange}
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
                renderOption={(props, option) => {
                  return (
                    <li {...props} key={option}>
                      {option}
                    </li>
                  )
                }}
                renderTags={(tagValue, getTagProps) => {
                  return tagValue.map((option, index) => (
                    <Chip
                      {...getTagProps({ index })}
                      key={option}
                      label={option}
                    />
                  ))
                }}
                renderInput={(params) => (
                  <TextField
                    {...params}
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
                name="password"
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
      )}
    </Modal>
  )
}
