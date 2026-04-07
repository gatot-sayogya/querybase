# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - generic [ref=e3]:
    - generic [ref=e4]:
      - generic [ref=e5]: Q
      - generic [ref=e6]: QueryBase
    - generic [ref=e7]:
      - heading "Welcome back" [level=1] [ref=e8]
      - paragraph [ref=e9]: Sign in to your account to continue
    - generic [ref=e10]:
      - generic [ref=e11]:
        - generic [ref=e12]: Username
        - textbox "Username" [ref=e13]:
          - /placeholder: Enter username
      - generic [ref=e14]:
        - generic [ref=e15]: Password
        - textbox "Password" [ref=e16]:
          - /placeholder: ••••••••
      - button "Sign in" [ref=e18] [cursor=pointer]
      - paragraph [ref=e20]: "Demo Credentials: admin / admin123"
    - paragraph [ref=e21]: Secure database access for your team
  - button "Open Next.js Dev Tools" [ref=e27] [cursor=pointer]:
    - img [ref=e28]
  - alert [ref=e31]
```