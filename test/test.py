import os
from os.path import dirname, join
from subprocess import check_output

import pytest
import requests
from syncloudlib.integration.hosts import add_host_alias
from syncloudlib.integration.installer import local_install, wait_for_installer

DIR = dirname(__file__)
TMP_DIR = '/tmp/syncloud'


@pytest.fixture(scope="session")
def module_setup(request, device, data_dir, platform_data_dir, app_dir, artifact_dir):
    def module_teardown():
        platform_log_dir = join(artifact_dir, 'platform_log')
        os.mkdir(platform_log_dir)
        device.scp_from_device('{0}/log/*'.format(platform_data_dir), platform_log_dir)

        device.run_ssh('top -bn 1 -w 500 -c > {0}/top.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ps auxfw > {0}/ps.log'.format(TMP_DIR), throw=False)
        device.run_ssh('netstat -nlp > {0}/netstat.log'.format(TMP_DIR), throw=False)
        device.run_ssh('journalctl > {0}/journalctl.log'.format(TMP_DIR), throw=False)
        device.run_ssh('cp -r /var/snap/amneziawg/current/config {0}/config.current'.format(TMP_DIR), throw=False)
        device.run_ssh('cp -r /snap/amneziawg/current/config {0}/config.app'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la /snap > {0}/snap.ls.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la {0}/ > {1}/app.ls.log'.format(app_dir, TMP_DIR), throw=False)
        device.run_ssh('ls -la {0}/ > {1}/data.ls.log'.format(data_dir, TMP_DIR), throw=False)

        app_log_dir = join(artifact_dir, 'log')
        os.mkdir(app_log_dir)
        device.scp_from_device('{0}/*'.format(TMP_DIR), app_log_dir)
        check_output('cp /etc/hosts {0}/hosts.log'.format(artifact_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)
    request.addfinalizer(module_teardown)


def test_start(module_setup, device, device_host, app, domain):
    add_host_alias(app, device_host, domain)
    device.run_ssh('date', retries=10)
    device.run_ssh('mkdir {0}'.format(TMP_DIR))


def test_activate_device(device):
    response = device.activate_custom()
    assert response.status_code == 200, response.text


def test_install_ca_cert(device):
    device.run_ssh('cp /var/snap/platform/current/syncloud.ca.crt /usr/local/share/ca-certificates/syncloud.crt')
    device.run_ssh('update-ca-certificates')


def test_install(app_archive_path, domain, device_session, device_password):
    local_install(domain, device_password, app_archive_path)
    wait_for_installer(device_session, domain, attempts=10)


def test_services(device):
    device.run_ssh('snap services amneziawg', throw=False)
    device.run_ssh('snap logs amneziawg -n 200', throw=False)
    device.run_ssh('ls -la /var/snap/amneziawg/current/ /var/snap/amneziawg/common/ /snap/amneziawg/current/', throw=False)
    device.run_ssh('cat /var/snap/amneziawg/current/config/awg0.conf', throw=False)


@pytest.mark.flaky(retries=10, delay=5)
def test_visible_through_platform(app_domain):
    response = requests.get('https://{0}'.format(app_domain), verify=False, allow_redirects=False)
    assert response.status_code in (200, 302), response.text


def test_server_interface_up(device):
    device.run_ssh('/snap/amneziawg/current/amneziawg-tools/bin/awg show awg0')


def test_storage_change_event(device):
    device.run_ssh('snap run amneziawg.storage-change > {0}/storage-change.log'.format(TMP_DIR))


def test_access_change_event(device):
    device.run_ssh('snap run amneziawg.access-change > {0}/access-change.log'.format(TMP_DIR))


def test_upgrade(app_archive_path, domain, device_password):
    local_install(domain, device_password, app_archive_path)


def test_remove(device, app):
    response = device.app_remove(app)
    assert response.status_code == 200, response.text


def test_reinstall(app_archive_path, domain, device_password):
    local_install(domain, device_password, app_archive_path)
