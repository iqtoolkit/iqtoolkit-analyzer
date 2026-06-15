// @ts-check
const {themes: prismThemes} = require('prism-react-renderer');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'iqtoolkit-analyzer',
  tagline: 'PostgreSQL health checking and performance tuning recommendations',
  favicon: 'img/favicon.ico',
  url: 'https://docs.iqtoolkit.ai',
  baseUrl: '/',
  organizationName: 'iqtoolkit',
  projectName: 'iqtoolkit-analyzer',
  trailingSlash: false,
  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: './sidebars.js',
          routeBasePath: '/',
          editUrl: 'https://github.com/iqtoolkit/iqtoolkit-analyzer/tree/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'iqtoolkit-analyzer',
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'tutorialSidebar',
            position: 'left',
            label: 'Docs',
          },
          {
            href: 'https://github.com/iqtoolkit/iqtoolkit-analyzer',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {label: 'Introduction', to: '/'},
              {label: 'Installation', to: '/installation'},
              {label: 'Usage', to: '/usage'},
            ],
          },
          {
            title: 'More',
            items: [
              {label: 'GitHub', href: 'https://github.com/iqtoolkit/iqtoolkit-analyzer'},
              {label: 'Releases', href: 'https://github.com/iqtoolkit/iqtoolkit-analyzer/releases'},
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} iqtoolkit. Built with Docusaurus.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
        additionalLanguages: ['bash', 'json'],
      },
    }),
};

module.exports = config;
