# QueryBase Frontend

Modern web interface for QueryBase built with Next.js 15, TypeScript, and Tailwind CSS.

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript 5
- **Styling**: Tailwind CSS 3
- **State Management**: Zustand
- **HTTP Client**: Axios
- **Testing**: Jest + React Testing Library
- **Editor**: Monaco Editor (SQL editor)

## Project Structure

```
web/
├── src/
│   ├── app/                    # Next.js app router pages
│   │   ├── login/             # Login page
│   │   ├── dashboard/         # Dashboard page
│   │   ├── layout.tsx         # Root layout
│   │   └── globals.css        # Global styles
│   ├── components/            # Reusable components
│   ├── lib/                   # Utility functions & API client
│   │   ├── api-client.ts      # Axios API client
│   │   └── utils.ts           # Utility functions
│   ├── stores/                # Zustand state management
│   │   └── auth-store.ts      # Authentication store
│   ├── types/                 # TypeScript type definitions
│   │   └── index.ts           # All types
│   ├── hooks/                 # Custom React hooks
│   └── __tests__/             # Test files
├── public/                    # Static assets
├── next.config.ts            # Next.js configuration
├── tailwind.config.ts        # Tailwind configuration
├── tsconfig.json             # TypeScript configuration
├── jest.config.js            # Jest configuration
└── package.json              # Dependencies

## Getting Started

### Prerequisites

- Node.js 18+ and npm

### Installation

```bash
cd web
npm install
```

### Configuration

Create a `.env.local` file:

```bash
cp .env.local.example .env.local
```

Edit `.env.local` if your API is running on a different port:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Development

Start the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Building for Production

```bash
npm run build
npm start
```

## Testing

Run tests:

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Generate coverage report
npm run test:coverage
```

## Features Implemented

### ✅ Phase 1: Foundation (Complete)

- [x] Project setup with Next.js 15, TypeScript, Tailwind CSS
- [x] Authentication flow with Zustand state management
- [x] Login page with form validation
- [x] Protected dashboard page
- [x] API client with all backend endpoints
- [x] JWT token handling
- [x] Error handling and display
- [x] Jest testing setup
- [x] Unit tests for utilities and API client

### 🚧 Phase 2: SQL Editor & Results (Next)

- [ ] Monaco editor integration
- [ ] Data source selector
- [ ] Query execution UI
- [ ] Results table with pagination
- [ ] Query status indicators
- [ ] Error display

### 📋 Phase 3: Approval Dashboard (Planned)

- [ ] Approval request list
- [ ] Approval detail view
- [ ] Approve/Reject buttons
- [ ] Comment discussion
- [ ] Transaction status display

### 📋 Phase 4: Admin Features (Planned)

- [ ] Data source management
- [ ] User management
- [ ] Group management
- [ ] Permission management

## API Endpoints

The frontend communicates with the backend API at `/api/v1/`:

- **Authentication**: `/api/v1/auth/*`
- **Queries**: `/api/v1/queries/*`
- **Approvals**: `/api/v1/approvals/*`
- **Data Sources**: `/api/v1/datasources/*`
- **Groups**: `/api/v1/groups/*`

See [CLAUDE.md](../CLAUDE.md) for complete API documentation.

## State Management

We use Zustand for client-side state management:

- `useAuthStore`: Authentication state (user, token, login/logout)
- Future stores: `useQueryStore`, `useDataSourceStore`, etc.

## Styling

We use Tailwind CSS for styling. The theme includes:

- Responsive design
- Dark mode support (via CSS variables)
- Consistent spacing and colors
- Utility-first approach

## Testing Strategy

- **Unit Tests**: Jest for utilities, hooks, and stores
- **Integration Tests**: React Testing Library for components
- **E2E Tests**: TBD (Playwright or Cypress)

## Default Credentials

For testing with the local backend:

- **Username**: `admin`
- **Password**: `admin123`

## Troubleshooting

### Port Already in Use

If port 3000 is already in use:

```bash
# Use a different port
PORT=3001 npm run dev
```

### API Connection Issues

If you can't connect to the API:

1. Ensure the backend is running: `make run-api` (from project root)
2. Check `NEXT_PUBLIC_API_URL` in `.env.local`
3. Check browser console for CORS errors

### Build Errors

If you encounter build errors:

```bash
# Clear Next.js cache
rm -rf .next

# Reinstall dependencies
rm -rf node_modules package-lock.json
npm install

# Rebuild
npm run build
```

## Development Tips

### Hot Reload

Next.js provides fast refresh. Changes to components will automatically reload in the browser.

### Debugging

- Use `console.log` for quick debugging
- Use React DevTools for component inspection
- Use browser DevTools for network debugging

### Code Style

- Use TypeScript for type safety
- Follow existing naming conventions
- Write tests for new utilities and hooks
- Document complex logic with comments

## Contributing

When adding new features:

1. Create a new branch: `git checkout -b feature/my-feature`
2. Follow the existing code structure
3. Write tests for new functionality
4. Ensure all tests pass: `npm test`
5. Build successfully: `npm run build`
6. Commit and push: `git commit -m "Add my feature"`

## Performance

- Next.js automatically code-splits by route
- Images are optimized with `next/image`
- Fonts are optimized with `next/font`
- Static pages are pre-rendered when possible

## Security

- JWT tokens stored in localStorage (consider httpOnly cookies for production)
- API routes protected with authentication middleware
- CSRF protection via SameSite cookies
- Environment variables for sensitive data

## License

Same as parent project (see parent directory).

---

**Last Updated:** January 28, 2025
**Status**: Phase 1 Complete ✅
