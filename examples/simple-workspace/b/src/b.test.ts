import { functionB } from './'

it('module b', () => {
  expect(functionB()).toEqual({
    b: 'b',
    c: {
      c: 'c',
      d: {
        d: 'd',
        e: { e: 'e' }
      },
      e: { e: 'e' }
    },
    d: {
      d: 'd',
      e: { e: 'e' }
    }
  })
})
