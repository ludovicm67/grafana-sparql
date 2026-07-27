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
