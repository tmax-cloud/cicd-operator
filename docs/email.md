# `Email`

This guide lets you know how to use `Email` feature.

* [Creating an Email step](#creating-an-email-step)

## Creating an `Email` step
Add following 'email' job in `IntegrationConfig`
```yaml
- name: email
  email:
     receivers:
        - sunghyun_kim3@tmax.co.kr
        - cqbqdd11519@gmail.com
     title: HTML Email title
     isHtml: true
     content: |
        <hr>
        <b>sdiofjsiodfjsdofj</b>
        <i>sdiofjsiodfjsdofj</i>
```
