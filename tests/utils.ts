import { PanelEditPage } from '@grafana/plugin-e2e';
import { expect, Locator, Page } from '@playwright/test';

/** Name of the provisioned datasource pointing at the local Oxigraph endpoint. */
export const LOCAL_DATASOURCE = 'SPARQL - Local';

/** Collapses every run of whitespace, so that rendered and source text compare equal. */
function normalize(text: string): string {
  return text.replace(/\s+/g, ' ').trim();
}

/**
 * Replaces the content of the Monaco based SPARQL editor.
 *
 * The query replaces the selection in a single edit: clearing the editor first
 * would briefly make the query text empty, and the re-render that follows can
 * put that empty value back into Monaco after the new query was typed.
 *
 * Monaco and the React state of the query editor still race with each other
 * often enough that the edit has to be verified, and retried when it did not
 * make it into the editor.
 */
export async function setSparqlQuery(page: Page, editor: Locator, query: string): Promise<void> {
  await expect(async () => {
    await editor.click();
    await page.keyboard.press('ControlOrMeta+KeyA');
    await page.keyboard.insertText(query);

    expect(normalize(await editor.innerText())).toContain(normalize(query));
  }).toPass({ timeout: 20000 });
}

/**
 * Sets the visualization of the panel being edited.
 *
 * Works around `panelEditPage.setVisualization` of @grafana/plugin-e2e, which
 * on Grafana >= 12.4 clicks the visualization picker toggle whenever it is
 * visible. Grafana 12.4.x keeps that toggle visible while the picker is open,
 * so the click closes the picker and the call times out. Here the toggle is
 * only clicked when the picker is not already open.
 */
export async function setVisualization(panelEditPage: PanelEditPage, visualization: string): Promise<void> {
  const { grafanaVersion, selectors } = panelEditPage.ctx;
  const [major, minor] = grafanaVersion.split('.').map(Number);
  if (major < 12 || (major === 12 && minor < 4)) {
    return panelEditPage.setVisualization(visualization);
  }

  const { components, constants } = selectors;
  const allVisualizationsTab = panelEditPage.getByGrafanaSelector(components.Tab.title(constants.Tab.title));
  const vizPickerToggle = panelEditPage.getByGrafanaSelector(components.PanelEditor.toggleVizPicker);
  const vizItem = panelEditPage.getByGrafanaSelector(components.PluginVisualization.item(visualization));
  const currentViz = panelEditPage.getByGrafanaSelector(components.PanelEditor.OptionsPane.header);

  await expect(async () => {
    if (!(await allVisualizationsTab.isVisible())) {
      await vizPickerToggle.click({ timeout: 2000 });
    }
    await allVisualizationsTab.click({ timeout: 2000 });
    await vizItem.click({ timeout: 2000 });
    await expect(currentViz).toHaveText(visualization, { timeout: 2000 });
  }).toPass({ timeout: 15000 });
}
