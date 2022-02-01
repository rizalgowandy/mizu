it('check', function () {
    cy.intercept('GET', 'http://localhost:8898/status/tap', { statusCode: 200 }).as('statusTap');
    cy.intercept('ws://localhost:8898/ws').as('ws');

    cy.visit(`http://localhost:8898/`);

    cy.wait('@statusTap');
    cy.wait('@ws')

    cy.get('.header').should('be.visible');
    cy.get('.TrafficPageHeader').should('be.visible');
    cy.get('.TrafficPage-ListContainer').should('be.visible');
    cy.get('.TrafficPage-Container').should('be.visible');
});
