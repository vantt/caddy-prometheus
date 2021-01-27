describe('Client Credential Test', () => {
    it('Does not do much!', () => {
        cy
            .request('/test-connect/client-credentials')
            .then((response) => {
                // response.body is automatically serialized into JSON
                expect(response.body).to.have.property('scope') // true
                expect(response.body).to.have.property('token_type') // true
                expect(response.body).to.have.property('access_token') // true
                expect(response.body).to.have.property('expires') // true
            })
    })
})

// $this->assertArrayHasKey('expires', $token);
// $this->assertArrayHasKey('access_token', $token);
// $this->assertArrayHasKey('token_type', $token);
// $this->assertSame('bearer', $token['token_type']);
// $this->assertNotEmpty($token['access_token']);
// $this->assertIsInt($token['expires']);