import {findLineAndCheck, getExpectedDetailsDict} from "../testHelpers/StatusBarHelper";

it('check', function () {
    const podName = 'httpbin', namespace = 'mizu-tests';

    cy.intercept('GET', 'http://localhost:8898/status/tap', { statusCode: 200 }).as('statusTap');

    cy.visit(`http://localhost:8898/`);

    cy.wait('@statusTap');
    findLineAndCheck(getExpectedDetailsDict(podName, namespace));

    cy.get('.header').should('be.visible');
    cy.get('.TrafficPageHeader').should('be.visible');
    cy.get('.TrafficPage-ListContainer').should('be.visible');
    cy.get('.TrafficPage-Container').should('be.visible');
});
