syntax on
filetype plugin indent on
set smartindent
set tabstop=4
set shiftwidth=4
set expandtab

" for solarized colorscheme
set termguicolors
syntax enable
set background=light
colorscheme solarized8

" for macvim ################### 
" for font
set guifont=Menlo\ Regular:h20

" remap insert mode ESC to home key jk
inoremap jk <ESC>

set number 
set noswapfile " disable swap file
set hlsearch " highlight all results"
set ignorecase " ignore case in search
set incsearch " show result as you type
set hidden " hides buffer instead of closing it

" for vim-go hightlighting
let g:go_highlight_types = 1
let g:go_highlight_fields = 1
let g:go_highlight_functions = 1
let g:go_highlight_function_calls = 1
let g:go_highlight_extra_types = 1
let g:go_highlight_operators = 1

let g:syntastic_go_checkers = ['golint', 'govet']
let g:syntastic_mode_map = { 'mode': 'active', 'passive_filetypes': ['go'] }
let g:go_list_type = "quickfix"

" for NERDTree
nmap ` :NERDTreeFind<CR>
nmap <C-A> :NERDTreeToggle<CR>
let NERDTreeShowHidden=1

" For vim
set clipboard=unnamed
set laststatus=2
set noshowmode

" For navigation
nmap <C-S> :b#<CR>

" for crtlP
let g:ctrlp_show_hidden = 1

" for lightline
let g:lightline = {
      \ 'colorscheme':'solarized', 
      \ 'active': {
      \   'left': [ [ 'mode', 'paste' ],
      \             [ 'gitbranch', 'readonly', 'filename', 'modified' ] ]
      \ },
      \ 'component_function': {
      \   'gitbranch': 'gitbranch#name'
      \ },
      \ 'component': {
      \	  'filename': '%n:%f'
      \	},
      \ }

" for remote debug
let g:go_debug_substitute_paths = [['/app/', '/Users/fp-od1120/Desktop/demo/pd-dine-in-order-service/']]

" for fugative
let g:github_enterprise_urls = ['https://github.com/deliveryhero']
