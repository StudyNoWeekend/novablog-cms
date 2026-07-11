export default [
  {
    path: '/setup',
    name: 'Setup',
    component: () => import('@/views/setup/SetupView.vue'),
    meta: { title: '系统初始化' },
  },
  {
    path: '/auth',
    children: [
      { path: 'login', name: 'Login', component: () => import('@/views/auth/LoginView.vue'), meta: { title: '登录' } },
    ],
  },
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        meta: { title: '工作台', icon: 'DashboardOutlined' },
        component: () => import('@/views/dashboard/DashboardView.vue'),
      },
      {
        path: 'articles',
        name: 'ArticleList',
        meta: { title: '文章管理', icon: 'FileTextOutlined' },
        component: () => import('@/views/article/ArticleListView.vue'),
      },
      {
        path: 'articles/create',
        name: 'ArticleCreate',
        meta: { title: '写文章', hidden: true },
        component: () => import('@/views/article/ArticleCreateView.vue'),
      },
      {
        path: 'articles/:id/edit',
        name: 'ArticleEdit',
        meta: { title: '编辑文章', hidden: true },
        component: () => import('@/views/article/ArticleEditView.vue'),
      },
      {
        path: 'media',
        name: 'MediaLibrary',
        meta: { title: '媒体库', icon: 'PictureOutlined' },
        component: () => import('@/views/media/MediaLibraryView.vue'),
      },
      {
        path: 'portfolios',
        name: 'Portfolios',
        meta: { title: '摄影作品集', icon: 'CameraOutlined' },
        component: () => import('@/views/portfolio/PortfolioListView.vue'),
      },
      {
        path: 'portfolios/:id/edit',
        name: 'PortfolioEdit',
        meta: { title: '编辑作品集', hidden: true },
        component: () => import('@/views/portfolio/PortfolioEditView.vue'),
      },
      {
        path: 'videos',
        name: 'Videos',
        meta: { title: '视频作品', icon: 'PlaySquareOutlined' },
        component: () => import('@/views/video/VideoManageView.vue'),
      },
      {
        path: 'travels',
        name: 'Travels',
        meta: { title: '旅行攻略', icon: 'CompassOutlined' },
        component: () => import('@/views/travel/TravelListView.vue'),
      },
      {
        path: 'travels/create',
        name: 'TravelCreate',
        meta: { title: '新建攻略', hidden: true },
        component: () => import('@/views/travel/TravelEditView.vue'),
      },
      {
        path: 'travels/:id/edit',
        name: 'TravelEdit',
        meta: { title: '编辑攻略', hidden: true },
        component: () => import('@/views/travel/TravelEditView.vue'),
      },
      {
        path: 'travels/:id',
        name: 'TravelDetail',
        meta: { title: '攻略详情', hidden: true },
        component: () => import('@/views/travel/TravelDetailView.vue'),
      },
      {
        path: 'playlists',
        name: 'Playlists',
        meta: { title: '音乐播放列表', icon: 'CustomerServiceOutlined' },
        component: () => import('@/views/playlist/PlaylistListView.vue'),
      },
      {
        path: 'comments',
        name: 'Comments',
        meta: { title: '评论管理', icon: 'MessageOutlined' },
        component: () => import('@/views/comment/CommentManageView.vue'),
      },
      {
        path: 'templates',
        name: 'Templates',
        meta: { title: '模板风格', icon: 'SkinOutlined' },
        component: () => import('@/views/template/TemplateView.vue'),
      },
      {
        path: 'profile',
        redirect: '/profile/info',
        component: () => import('@/components/layout/SettingsLayout.vue'),
        children: [
          {
            path: 'info',
            name: 'Profile',
            meta: { title: '个人资料' },
            component: () => import('@/views/user/ProfileView.vue'),
          },
          {
            path: 'storage',
            name: 'StorageConfig',
            meta: { title: '对象存储' },
            component: () => import('@/views/storage/StorageConfigView.vue'),
          },
        ],
      },
    ],
  },
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/error/NotFoundView.vue') },
]