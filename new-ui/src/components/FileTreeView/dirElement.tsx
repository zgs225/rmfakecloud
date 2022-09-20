/* eslint-disable tailwindcss/no-custom-classname, @typescript-eslint/no-unused-vars */

import { FolderIcon } from '@heroicons/react/outline'
import { ErrorMessage, Field, Form, Formik } from 'formik'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { PulseLoader } from 'react-spinners'
import { toast } from 'react-toastify'
import * as Yup from 'yup'

import { createFolder } from '../../api/document'
import { inputClassName } from '../../utils/form'
import { HashDoc } from '../../utils/models'

import { HashDocElementProp } from './props'

export default function DirElement(params: HashDocElementProp) {
  const {
    doc,
    onClickDoc,
    onDocEditingDiscard,
    onDocRenamed,
    onFolderCreated,
    onFolderCreationDiscarded,
    className,
    ...remainParams
  } = params
  const [unmountForm, setUnmountForm] = useState(false)
  const { t } = useTranslation()

  const validationSchema = Yup.object().shape({
    name: Yup.string().required(t('documents.new_folder_form.name.required'))
  })

  const { preMode } = doc
  let { mode } = doc

  if (!mode) {
    mode = 'display'
  }
  const creatingToDisplay = mode === 'display' && preMode === 'creating'

  const formFadeout = () => {
    setTimeout(() => {
      setUnmountForm(true)
    }, 500)
  }

  if (creatingToDisplay) {
    formFadeout()
  }

  const formDom = (
    <Formik
      initialValues={{ name: doc.name }}
      validationSchema={validationSchema}
      onSubmit={(values, { setSubmitting }) => {
        setSubmitting(true)

        createFolder(values.name)
          .then((response) => {
            const newDoc = { ...(response.data as HashDoc), children: [] }

            toast.success(t('notifications.folder_created'))
            onFolderCreated && onFolderCreated(newDoc, remainParams.index)

            return 'ok'
          })
          .catch((err) => {
            throw err
          })
          .finally(() => {
            setSubmitting(false)
          })
      }}
    >
      {({ isSubmitting, errors, touched }) => (
        <Form
          className={`w-full overflow-hidden ${
            creatingToDisplay ? 'animate-roll-up' : 'animate-roll-down'
          }`}
        >
          <div className="mb-4">
            <label className="mb-2 block font-bold text-neutral-400">
              {t('documents.new_folder_form.name.label')}
            </label>
            <Field
              autoFocus={true}
              className={inputClassName(errors.name && touched.name)}
              name="name"
              type="text"
            />
            <ErrorMessage
              className="mt-2 text-xs text-red-600"
              component="div"
              name="name"
            />
          </div>
          <div className="flex">
            <button
              className="mr-2 w-full basis-1/2 rounded border border-slate-600 py-3 font-bold text-neutral-200 focus:outline-none"
              type="button"
              onClick={(e) => {
                e.stopPropagation()
                setUnmountForm(false)
                onFolderCreationDiscarded && onFolderCreationDiscarded(doc, remainParams.index)
              }}
            >
              {t('documents.new_folder_form.cancel-btn')}
            </button>
            <button
              className="w-full basis-1/2 rounded bg-blue-700 py-3 font-bold text-neutral-200 hover:bg-blue-600 focus:outline-none disabled:bg-blue-500"
              disabled={isSubmitting}
              type="submit"
            >
              {isSubmitting ? (
                <PulseLoader
                  color="#e5e5e5"
                  cssOverride={{ lineHeight: 0, padding: '6px 0' }}
                  size={8}
                  speedMultiplier={0.8}
                />
              ) : (
                t('documents.new_folder_form.submit-btn')
              )}
            </button>
          </div>
        </Form>
      )}
    </Formik>
  )

  let innerDom: JSX.Element

  if (mode === 'creating') {
    innerDom = formDom
  } else {
    innerDom = (
      <>
        <FolderIcon className="top-[-1px] mr-2 h-6 w-6 shrink-0" />
        <p className="max-w-[calc(100%-28px)] overflow-hidden text-ellipsis whitespace-nowrap leading-6">
          {doc.name}
        </p>
      </>
    )
  }

  return (
    <>
      <div
        className={`flex cursor-pointer py-6 ${className || ''}`}
        {...remainParams}
        onClick={(e) => {
          if (doc.mode === 'creating') {
            return
          }

          e.preventDefault()
          e.stopPropagation()

          onClickDoc && onClickDoc(doc)
        }}
      >
        {innerDom}
      </div>
      {creatingToDisplay && !unmountForm ? formDom : <></>}
    </>
  )
}
