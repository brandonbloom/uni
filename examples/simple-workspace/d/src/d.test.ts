import { functionD } from './'

it('module d', () => {
  expect(functionD()).toEqual({
    d: 'd',
    e: { e: 'e' }
  })
})
