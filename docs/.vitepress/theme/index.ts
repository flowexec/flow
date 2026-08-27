import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import { enhanceAppWithTabs } from 'vitepress-plugin-tabs/client'
import Layout from './Layout.vue'
import ExamplesGallery from './ExamplesGallery.vue'
import Icon from './Icon.vue'
import SectionHead from './SectionHead.vue'
import NumberedSection from './NumberedSection.vue'
import CardGrid from './CardGrid.vue'
import Card from './Card.vue'
import NowStrip from './NowStrip.vue'
import './custom.css'

export default {
    extends: DefaultTheme,
    Layout,
    enhanceApp({ app }) {
        enhanceAppWithTabs(app)
        app.component('ExamplesGallery', ExamplesGallery)
        app.component('Icon', Icon)
        app.component('SectionHead', SectionHead)
        app.component('NumberedSection', NumberedSection)
        app.component('CardGrid', CardGrid)
        app.component('Card', Card)
        app.component('NowStrip', NowStrip)
    },
} satisfies Theme
