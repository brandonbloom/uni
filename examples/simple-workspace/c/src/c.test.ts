import { functionC } from './'

it('module c', () => {
  expect(functionC()).toEqual({
    c: 'c',
    d: {
      d: 'd',
      e: { e: 'e' }
    },
    e: { e: 'e' }
  })
})
