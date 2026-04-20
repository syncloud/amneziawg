import pytest
from subprocess import check_output

import requests
from syncloudlib.integration.hosts import add_host_alias
from syncloudlib.integration.installer import local_install, wait_for_installer

TMP_DIR = '/tmp/syncloud'


@pytest.fixture(scope="session")
def module_setup(request, device, artifact_dir):
    def module_teardown():
        device.run_ssh('journalctl > {0}/refresh.journalctl.log'.format(TMP_DIR), throw=False)
        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, app, device_host, domain, device):
    add_host_alias(app, device_host, domain)
    device.activated()
    device.run_ssh('rm -rf {0}'.format(TMP_DIR), throw=False)
    device.run_ssh('mkdir {0}'.format(TMP_DIR), throw=False)


# TODO: enable once the app has a released version in the Syncloud store.
# def test_install_released(app_archive_path_latest, domain, device_session, device_password):
#     local_install(domain, device_password, app_archive_path_latest)
#     wait_for_installer(device_session, domain, attempts=10)


def test_upgrade_to_new(app_archive_path, domain, device_session, device_password):
    local_install(domain, device_password, app_archive_path)
    wait_for_installer(device_session, domain, attempts=10)


@pytest.mark.flaky(retries=10, delay=5)
def test_visible_through_platform(app_domain):
    response = requests.get('https://{0}'.format(app_domain), verify=False, allow_redirects=False)
    assert response.status_code in (200, 302), response.text
