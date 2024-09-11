import { describe, it, expect } from 'vitest'

import { mount } from '@vue/test-utils'
import navBar from '../navBar.vue'

describe('navBar', () => {
  it('renders properly', () => {
    const wrapper = mount(navBar, { props: { msg: 'Hello Vitest' } })
    expect(wrapper.text()).toContain('Hello Vitest')
  })
})
