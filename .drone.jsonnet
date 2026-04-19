local name = 'amneziawg';

local go = '1.26';
local node = '20.18.1';
local nginx = '1.29.3-alpine3.22';
local alpine = '3.22.2';
local debian = 'bookworm-slim';
local python = '3.12-slim-bookworm';
local playwright = 'v1.59.1-jammy';
local browser = 'chrome';
local dind = '20.10.21-dind';

// Syncloud platform image tags. Bookworm tracks the latest release on
// Docker Hub at the time of editing — bump this when a newer tag ships
// at https://hub.docker.com/r/syncloud/platform-bookworm-amd64/tags.
// Buster stays pinned because newer platform builds dropped buster.
local platform = '26.04.3';
local platform_buster = '25.02';

local distro_default = 'bookworm';
local distros = ['bookworm', 'buster'];

local deployer = 'https://github.com/syncloud/store/releases/download/4/syncloud-release';

// Upstream versions — pin here so bumps show up as a CI diff. Build
// scripts accept the version as a positional argument.
local amneziawg_go_version = '1.1.0';
local amneziawg_tools_version = '1.0.20250706';

local platform_image(distro, arch) =
  'syncloud/platform-' + distro + '-' + arch + ':' +
  (if distro == 'buster' then platform_buster else platform);

local build(arch, test_ui) = [{
  kind: 'pipeline',
  type: 'docker',
  name: arch,
  platform: { os: 'linux', arch: arch },

  steps: [
    {
      name: 'version',
      image: 'debian:' + debian,
      commands: ['echo $DRONE_BUILD_NUMBER > version'],
    },
    {
      name: 'nginx',
      image: 'nginx:' + nginx,
      commands: ['./nginx/build.sh'],
    },
    {
      name: 'nginx test',
      image: platform_image(distro_default, arch),
      commands: ['./nginx/test.sh'],
    },
    {
      name: 'amneziawg-go',
      image: 'golang:' + go,
      commands: ['./amneziawg-go/build.sh ' + amneziawg_go_version],
    },
    {
      name: 'amneziawg-tools',
      image: 'alpine:' + alpine,
      commands: [
        'apk add --no-cache build-base linux-headers bash',
        './amneziawg-tools/build.sh ' + amneziawg_tools_version,
      ],
    },
    {
      name: 'cli',
      image: 'golang:' + go,
      commands: ['./cli/build.sh'],
    },
    {
      name: 'backend',
      image: 'golang:' + go,
      commands: ['./backend/build.sh'],
    },
    {
      name: 'web',
      image: 'node:' + node,
      commands: ['./web/build.sh'],
    },
    {
      name: 'package',
      image: 'debian:' + debian,
      commands: [
        'VERSION=$(cat version)',
        './package.sh ' + name + ' $VERSION',
      ],
    },
  ] + [
    {
      name: 'test ' + distro,
      image: 'python:' + python,
      commands: [
        './ci/integration.sh ' + name + ' ' + distro + ' $DRONE_BUILD_NUMBER',
      ],
    }
    for distro in distros
  ] + (if test_ui then [
         {
           name: 'test-ui-desktop',
           image: 'mcr.microsoft.com/playwright:' + playwright,
           commands: [
             './ci/ui.sh desktop ' + name + ' ' + distro_default + ' $DRONE_BUILD_NUMBER',
           ],
         },
         {
           name: 'test-ui-mobile',
           image: 'mcr.microsoft.com/playwright:' + playwright,
           commands: [
             './ci/ui.sh mobile ' + name + ' ' + distro_default + ' $DRONE_BUILD_NUMBER',
           ],
         },
       ] else []) + (if arch == 'amd64' then [
                      {
                        name: 'test-upgrade',
                        image: 'python:' + python,
                        privileged: true,
                        commands: [
                          './ci/upgrade.sh ' + name + ' ' + distro_default + ' $DRONE_BUILD_NUMBER ' + browser,
                        ],
                      },
                    ] else []) + [
    {
      name: 'upload',
      image: 'debian:' + debian,
      environment: {
        AWS_ACCESS_KEY_ID: { from_secret: 'AWS_ACCESS_KEY_ID' },
        AWS_SECRET_ACCESS_KEY: { from_secret: 'AWS_SECRET_ACCESS_KEY' },
        SYNCLOUD_TOKEN: { from_secret: 'SYNCLOUD_TOKEN' },
      },
      commands: ['./ci/upload.sh ' + arch + ' ' + deployer + ' $DRONE_BRANCH'],
      when: { branch: ['stable', 'master'], event: ['push'] },
    },
    {
      name: 'promote',
      image: 'debian:' + debian,
      environment: {
        AWS_ACCESS_KEY_ID: { from_secret: 'AWS_ACCESS_KEY_ID' },
        AWS_SECRET_ACCESS_KEY: { from_secret: 'AWS_SECRET_ACCESS_KEY' },
        SYNCLOUD_TOKEN: { from_secret: 'SYNCLOUD_TOKEN' },
      },
      commands: ['./ci/promote.sh ' + arch + ' ' + name + ' ' + deployer],
      when: { branch: ['stable'], event: ['push'] },
    },
    {
      name: 'artifact',
      image: 'appleboy/drone-scp:1.6.4',
      settings: {
        host: { from_secret: 'artifact_host' },
        username: 'artifact',
        key: { from_secret: 'artifact_key' },
        timeout: '2m',
        command_timeout: '2m',
        target: '/home/artifact/repo/' + name + '/${DRONE_BUILD_NUMBER}-' + arch,
        source: 'artifact/*',
        strip_components: 1,
      },
      when: { status: ['failure', 'success'], event: ['push'] },
    },
  ],

  trigger: { event: ['push', 'pull_request'] },

  services: [
    {
      name: name + '.' + distro + '.com',
      image: platform_image(distro, arch),
      privileged: true,
      volumes: [
        { name: 'dbus', path: '/var/run/dbus' },
        { name: 'dev', path: '/dev' },
      ],
    }
    for distro in distros
  ],

  volumes: [
    { name: 'dbus', host: { path: '/var/run/dbus' } },
    { name: 'dev', host: { path: '/dev' } },
    { name: 'shm', temp: {} },
    { name: 'videos', temp: {} },
    { name: 'dockersock', temp: {} },
  ],
}];

// armhf (arm/v7) is attempted as a third target because older devices
// still run on it. If the arm pipeline starts failing consistently
// (e.g. upstream drops arm/v7 support), drop this line — amd64 + arm64
// coverage is the non-negotiable minimum.
build('amd64', true) +
build('arm64', false) +
build('arm', false)
