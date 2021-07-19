import { functionA } from './'

it('module a', () => {
  expect(functionA()).toEqual({
    a: 'a',
    b: {
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
    }
  })
})
