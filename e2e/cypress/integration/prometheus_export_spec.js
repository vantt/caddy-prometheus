describe('Caddy correctly setup', () => {
    it('Can access /path1', () => {
        cy
            .request('/path1')
            .then((response) => {
                expect(response.status).to.equal(200)
                expect(response.body).to.equal('Hello path1')
            })
    })

    it('Can access /path2', () => {
        cy
            .request('/path2')
            .then((response) => {
                // response.body is automatically serialized into JSON
                expect(response.status).to.equal(200)
                expect(response.body).to.equal('Hello path2')
            })
    })
})


describe('Caddy correctly expose metrics', () => {
    it('Metric show correctly', () => {
        cy
            .request(Cypress.env('METRICS_URL'))
            .then((response) => {
                // response.body is automatically serialized into JSON
                expect(response.status).to.equal(200)
                expect(response.body).to.contains('caddy2_http_request_count_total{family="1",host="caddy",proto="1.1",route_name="/path1",server="Caddy"} 1')
                expect(response.body).to.contains('caddy2_http_request_count_total{family="1",host="caddy",proto="1.1",route_name="/path2",server="Caddy"} 1')
            })
    })
})