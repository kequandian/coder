document.addEventListener('DOMContentLoaded', () => {
    const chatMessages = document.getElementById('chatMessages');
    const chatForm = document.getElementById('chatForm');
    const userInput = document.getElementById('userInput');
    const sendButton = document.getElementById('sendButton');
    const newChatBtn = document.getElementById('newChatBtn');
    const conversationList = document.getElementById('conversationList');

    // 确保样式表加载
    function ensureStylesLoaded() {
        const stylesLink = document.querySelector('link[href*="style.css"]');
        if (!stylesLink) {
            const link = document.createElement('link');
            link.rel = 'stylesheet';
            link.href = 'style.css';
            document.head.appendChild(link);
            console.log('已添加style.css样式表');
        }
    }
    
    // 确保样式表加载
    ensureStylesLoaded();

    // 创建HTML预览模态框
    createHtmlPreviewModal();

    // 初始化Markdown-it解析器
    const md = window.markdownit({
        html: false,        // 禁用HTML标签
        breaks: true,       // 允许换行
        linkify: true,      // 自动转换URL为链接
        highlight: function (str, lang) {
            if (lang && hljs.getLanguage(lang)) {
                try {
                    return hljs.highlight(str, { language: lang }).value;
                } catch (__) {}
            }
            return ''; // 使用默认转义
        }
    });

    // 添加KaTeX渲染支持
    if (window.katex && window.texmath) {
        const tm = window.texmath.use(window.katex);
        md.use(tm, { engine: window.katex, delimiters: 'dollars' });
    }

    // 保存所有会话数据
    let conversations = {};
    
    // 当前会话ID
    let currentConversationId = null;
    
    // 初始化
    function initialize() {
        // 从本地存储加载会话
        loadConversationsFromStorage();
        
        // 如果没有会话，创建一个新会话
        if (Object.keys(conversations).length === 0) {
            createNewConversation();
        } else {
            // 加载上次活跃的会话或第一个会话
            const lastActiveId = localStorage.getItem('lastActiveConversation');
            if (lastActiveId && conversations[lastActiveId]) {
                loadConversation(lastActiveId);
            } else {
                loadConversation(Object.keys(conversations)[0]);
            }
        }
        
        // 渲染会话列表
        renderConversationList();
    }
    
    // 从本地存储加载会话
    function loadConversationsFromStorage() {
        const savedConversations = localStorage.getItem('conversations');
        if (savedConversations) {
            try {
                conversations = JSON.parse(savedConversations);
            } catch (e) {
                console.error('Failed to load conversations:', e);
                conversations = {};
            }
        }
    }
    
    // 保存会话到本地存储
    function saveConversationsToStorage() {
        localStorage.setItem('conversations', JSON.stringify(conversations));
        localStorage.setItem('lastActiveConversation', currentConversationId);
    }
    
    // 创建新会话
    function createNewConversation() {
        const id = generateId();
        currentConversationId = id;
        
        conversations[id] = {
            id: id,
            title: '新聊天',
            messages: [],
            createdAt: new Date().toISOString()
        };
        
        // 清空聊天区域
        chatMessages.innerHTML = '';
        addMessage('欢迎使用智能聊天系统！请在下方输入你的问题。', 'system');
        
        // 更新侧边栏
        renderConversationList();
        saveConversationsToStorage();
    }
    
    // 加载会话
    function loadConversation(id) {
        if (!conversations[id]) return;
        
        currentConversationId = id;
        
        // 清空聊天区域
        chatMessages.innerHTML = '';
        
        // 加载会话消息
        const conversation = conversations[id];
        messageHistory = [...conversation.messages];
        
        if (messageHistory.length === 0) {
            // 如果是空会话，显示欢迎消息
            addMessage('欢迎使用智能聊天系统！请在下方输入你的问题。', 'system');
        } else {
            // 渲染所有消息
            messageHistory.forEach(msg => {
                addMessageToUI(msg.content, msg.role);
            });
        }
        
        // 更新侧边栏选中状态
        updateActiveConversation(id);
        
        saveConversationsToStorage();
    }
    
    // 渲染会话列表
    function renderConversationList() {
        conversationList.innerHTML = '';
        
        // 按创建时间倒序排列
        const sortedIds = Object.keys(conversations).sort((a, b) => {
            return new Date(conversations[b].createdAt) - new Date(conversations[a].createdAt);
        });
        
        sortedIds.forEach(id => {
            const conversation = conversations[id];
            const isActive = id === currentConversationId;
            
            const itemEl = document.createElement('div');
            itemEl.className = `conversation-item ${isActive ? 'active' : ''}`;
            itemEl.dataset.id = id;
            
            // 计算相对时间
            const relativeTime = getRelativeTime(conversation.createdAt);
            
            itemEl.innerHTML = `
                <div class="conversation-icon">
                    <i class="fas fa-comments"></i>
                </div>
                <div class="conversation-info">
                    <div class="conversation-title">${conversation.title}</div>
                    <div class="conversation-time">${relativeTime}</div>
                </div>
                <div class="conversation-delete" title="删除对话">
                    <i class="fas fa-times"></i>
                </div>
            `;
            
            // 添加点击会话的事件监听器
            itemEl.addEventListener('click', (e) => {
                // 如果点击的是删除按钮，则删除会话
                if (e.target.closest('.conversation-delete')) {
                    e.stopPropagation(); // 阻止冒泡
                    deleteConversation(id);
                } else {
                    loadConversation(id);
                }
            });
            
            conversationList.appendChild(itemEl);
        });
    }
    
    // 计算相对时间
    function getRelativeTime(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now - date;
        const diffSeconds = Math.floor(diffMs / 1000);
        const diffMinutes = Math.floor(diffSeconds / 60);
        const diffHours = Math.floor(diffMinutes / 60);
        const diffDays = Math.floor(diffHours / 24);
        
        if (diffDays > 0) {
            return `${diffDays}天前`;
        } else if (diffHours > 0) {
            return `${diffHours}小时前`;
        } else if (diffMinutes > 0) {
            return `${diffMinutes}分钟前`;
        } else {
            return "刚刚";
        }
    }
    
    // 删除会话
    function deleteConversation(id) {
        if (!conversations[id]) return;
        
        // 删除前先确认
        if (!confirm("确定要删除这个对话吗？此操作不可撤销。")) {
            return;
        }
        
        // 删除会话
        delete conversations[id];
        saveConversationsToStorage();
        
        // 如果删除的是当前会话，则加载另一个会话
        if (id === currentConversationId) {
            const remainingIds = Object.keys(conversations);
            if (remainingIds.length > 0) {
                loadConversation(remainingIds[0]);
            } else {
                createNewConversation();
            }
        }
        
        // 重新渲染会话列表
        renderConversationList();
    }
    
    // 更新会话列表的选中状态
    function updateActiveConversation(id) {
        document.querySelectorAll('.conversation-item').forEach(item => {
            if (item.dataset.id === id) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        });
    }
    
    // 更新会话标题
    function updateConversationTitle(id, title) {
        if (!conversations[id]) return;
        
        conversations[id].title = title;
        saveConversationsToStorage();
        renderConversationList();
    }
    
    // Generate a random ID
    function generateId() {
        return Math.random().toString(36).substring(2, 15);
    }
    
    // Keep track of the message history
    let messageHistory = [];

    // Add a typing indicator
    function addTypingIndicator() {
        const div = document.createElement('div');
        div.className = 'message assistant typing-indicator';
        div.innerHTML = `
            <div class="message-content">
                <p>正在思考<span class="typing"></span></p>
            </div>
        `;
        div.id = 'typing-indicator';
        chatMessages.appendChild(div);
        chatMessages.scrollTop = chatMessages.scrollHeight;
        return div;
    }

    // Remove typing indicator
    function removeTypingIndicator() {
        const indicator = document.getElementById('typing-indicator');
        if (indicator) {
            indicator.remove();
        }
    }

    // Add a message to the UI without affecting message history
    function addMessageToUI(content, role) {
        const div = document.createElement('div');
        div.className = `message ${role}`;
        
        let formattedContent = content;
        
        // 使用markdown-it渲染内容
        if (role === 'assistant' || role === 'system') {
            console.log('content:', content);
            formattedContent = md.render(formattedContent);
        } else {
            // 用户消息只做简单的转义和换行处理
            formattedContent = content
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#039;')
                .replace(/\n/g, '<br>');
        }
        
        div.innerHTML = `
            <div class="message-content">
                ${role === 'user' ? `<p>${formattedContent}</p>` : formattedContent}
            </div>
        `;
        
        chatMessages.appendChild(div);
        chatMessages.scrollTop = chatMessages.scrollHeight;
        
        // 如果有代码块，应用语法高亮
        if (role === 'assistant' || role === 'system') {
            div.querySelectorAll('pre code').forEach((block) => {
                hljs.highlightElement(block);
                
                // 获取语言并添加到pre元素上
                const language = block.className.split('-')[1];
                if (language) {
                    block.parentElement.setAttribute('data-language', language);
                    
                    // 如果是HTML代码块，添加预览按钮
                    if (language === 'html') {
                        // 检查是否已经添加了预览按钮
                        const preBlock = block.parentElement;
                        if (!preBlock.querySelector('.html-preview-btn')) {
                            // 创建预览按钮
                            const previewBtn = document.createElement('button');
                            previewBtn.className = 'html-preview-btn';
                            previewBtn.textContent = '预览HTML';
                            previewBtn.onclick = function(e) {
                                e.preventDefault();
                                e.stopPropagation();
                                console.log('预览按钮被点击');
                                showHtmlPreview(block.textContent);
                            };
                            
                            // 添加按钮到pre元素
                            preBlock.appendChild(previewBtn);
                        }
                    }
                }
            });
        }
    }
    
    // Add a message with history tracking
    function addMessage(content, role) {
        removeTypingIndicator();
        
        // Add to UI
        addMessageToUI(content, role);
        
        // Add to history and current conversation (except system messages)
        if (role !== 'system') {
            messageHistory.push({
                role: role,
                content: content
            });
            
            // 保存到当前会话
            if (currentConversationId && conversations[currentConversationId]) {
                conversations[currentConversationId].messages = [...messageHistory];
                
                // 如果是用户的第一条消息，使用它作为会话标题
                const conversation = conversations[currentConversationId];
                if (role === 'user' && conversation.messages.filter(m => m.role === 'user').length === 1) {
                    // 截取不超过30个字符的标题
                    const title = content.substring(0, 30) + (content.length > 30 ? '...' : '');
                    console.log('更新对话标题为:', title);
                    updateConversationTitle(currentConversationId, title);
                }
                
                saveConversationsToStorage();
            }
        }
    }

    // Send a message to the API
    async function sendMessage(message) {
        // Add user message to UI
        addMessage(message, 'user');
        
        // Disable input during processing
        userInput.disabled = true;
        sendButton.disabled = true;
        
        // Show typing indicator
        addTypingIndicator();
        
        try {
            // 只保留最后四条消息   
            messages = messageHistory.slice(-1);
            const response = await fetch('/v1/chat/completions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    id: generateId(),
                    conversation_id: currentConversationId,
                    messages: messages,
                    stream: true
                })
            });

            if (!response.ok) {
                throw new Error(`API responded with status ${response.status}`);
            }

            // Handle streaming response
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';
            let assistantResponse = '';
            
            removeTypingIndicator();
            
            // Create a message div for the assistant response
            const assistantDiv = document.createElement('div');
            assistantDiv.className = 'message assistant';
            assistantDiv.innerHTML = `
                <div class="message-content">
                    <div class="markdown-content"></div>
                </div>
            `;
            chatMessages.appendChild(assistantDiv);
            const assistantContent = assistantDiv.querySelector('.markdown-content');
            
            while (true) {
                const { value, done } = await reader.read();
                if (done) break;
                
                // Decode the stream
                buffer += decoder.decode(value, { stream: true });
                
                // Process all complete SSE events
                const lines = buffer.split('\n\n');
                buffer = lines.pop() || '';
                
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.substring(6);
                        if (data === '[DONE]') {
                            // All done
                            break;
                        }
                        
                        try {
                            const parsed = JSON.parse(data);
                            const content = parsed.choices[0]?.delta?.content || '';
                            if (content) {
                                assistantResponse += content;
                                
                                // 使用Markdown渲染当前响应
                                let formattedResponse = assistantResponse;
                                assistantContent.innerHTML = md.render(formattedResponse);
                                
                                // 应用语法高亮
                                assistantDiv.querySelectorAll('pre code').forEach((block) => {
                                    hljs.highlightElement(block);
                                    
                                    // 获取语言并添加到pre元素上
                                    const language = block.className.split('-')[1];
                                    if (language) {
                                        block.parentElement.setAttribute('data-language', language);
                                    }
                                });
                                
                                // 为HTML代码块添加预览按钮
                                addHtmlPreviewButtons();
                                
                                chatMessages.scrollTop = chatMessages.scrollHeight;
                            }
                        } catch (e) {
                            console.error('Error parsing SSE data:', e);
                        }
                    }
                }
            }
            
            // Add final message to history
            messageHistory.push({
                role: 'assistant',
                content: assistantResponse
            });
            
            // 保存到当前会话
            if (currentConversationId && conversations[currentConversationId]) {
                conversations[currentConversationId].messages = [...messageHistory];
                
                // 检查是否需要更新对话标题 (如果仍然是默认标题"新聊天")
                const conversation = conversations[currentConversationId];
                if (conversation.title === '新聊天' && messageHistory.length >= 2) {
                    // 获取用户的第一条消息
                    const firstUserMessage = messageHistory.find(m => m.role === 'user');
                    if (firstUserMessage) {
                        const title = firstUserMessage.content.substring(0, 30) + 
                                     (firstUserMessage.content.length > 30 ? '...' : '');
                        updateConversationTitle(currentConversationId, title);
                    }
                }
                
                saveConversationsToStorage();
            }
            
        } catch (error) {
            console.error('Error:', error);
            addMessage('发生错误，请重试。' + error.message, 'system');
        } finally {
            // Re-enable input
            userInput.disabled = false;
            sendButton.disabled = false;
            userInput.focus();
        }
    }

    // 绑定新建聊天按钮事件
    newChatBtn.addEventListener('click', () => {
        createNewConversation();
    });

    // Handle form submission
    chatForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const message = userInput.value.trim();
        if (message) {
            sendMessage(message);
            userInput.value = '';
        }
    });

    // Allow sending with Enter (but not with Shift+Enter)
    userInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            chatForm.dispatchEvent(new Event('submit'));
        }
    });
    
    // 初始化应用
    initialize();

    // HTML预览功能
    function addHtmlPreviewButtons() {
        document.querySelectorAll('pre code.language-html').forEach((block) => {
            // 检查是否已经添加了预览按钮
            const preBlock = block.parentElement;
            if (preBlock.querySelector('.html-preview-btn')) return;
            
            // 创建预览按钮
            const previewBtn = document.createElement('button');
            previewBtn.className = 'html-preview-btn';
            previewBtn.textContent = '预览HTML';
            previewBtn.onclick = function(e) {
                e.preventDefault();
                e.stopPropagation();
                console.log('预览按钮被点击');
                showHtmlPreview(block.textContent);
            };
            
            // 添加按钮到pre元素
            preBlock.appendChild(previewBtn);
        });
    }

    function createHtmlPreviewModal() {
        // 检查是否已存在模态框
        if (document.getElementById('html-preview-modal')) {
            return document.getElementById('html-preview-modal');
        }
        
        // 创建模态框容器
        const modal = document.createElement('div');
        modal.id = 'html-preview-modal';
        modal.className = 'modal';
        modal.style.display = 'none'; // 确保初始状态为隐藏
        
        // 创建模态框内容
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h2>HTML预览</h2>
                    <div class="modal-controls">
                        <button id="fullscreen-btn" title="全屏预览"><i class="fas fa-expand"></i></button>
                        <button id="close-preview-btn" title="关闭预览"><i class="fas fa-times"></i></button>
                    </div>
                </div>
                <div class="modal-body">
                    <iframe id="html-preview-iframe" sandbox="allow-same-origin allow-scripts"></iframe>
                </div>
            </div>
        `;
        
        // 添加到文档中
        document.body.appendChild(modal);
        
        // 添加事件监听器
        document.getElementById('close-preview-btn').addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            modal.style.display = 'none';
        });
        
        document.getElementById('fullscreen-btn').addEventListener('click', function(e) {
            e.preventDefault();
            e.stopPropagation();
            const iframe = document.getElementById('html-preview-iframe');
            if (iframe.requestFullscreen) {
                iframe.requestFullscreen();
            } else if (iframe.webkitRequestFullscreen) {
                iframe.webkitRequestFullscreen();
            } else if (iframe.msRequestFullscreen) {
                iframe.msRequestFullscreen();
            }
        });
        
        // 点击模态框外部时关闭
        modal.addEventListener('click', function(event) {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
        
        return modal;
    }

    function showHtmlPreview(htmlCode) {
        // 确保模态框存在
        const modal = createHtmlPreviewModal() || document.getElementById('html-preview-modal');
        
        // 显示模态框
        modal.style.display = 'block';
        
        // 将HTML代码加载到iframe中
        const iframe = document.getElementById('html-preview-iframe');
        
        try {
            // 使用新的iframe src方法加载内容，避免变量重复声明问题
            const htmlTemplate = `
                <!DOCTYPE html>
                <html>
                <head>
                    <meta charset="UTF-8">
                    <meta name="viewport" content="width=device-width, initial-scale=1.0">
                    <style>
                        body { margin: 0; padding: 10px; font-family: Arial, sans-serif; }
                        img { max-width: 100%; height: auto; } /* 添加图片响应式样式 */
                    </style>
                </head>
                <body>
                    ${htmlCode}
                </body>
                </html>
            `;
            
            // 使用Blob和createObjectURL创建一个临时URL来加载HTML
            const blob = new Blob([htmlTemplate], { type: 'text/html' });
            const url = URL.createObjectURL(blob);
            
            // 设置iframe的src
            iframe.src = url;
            
            // 设置onload事件，确保加载完成后释放blob URL
            iframe.onload = function() {
                URL.revokeObjectURL(url);
                modal.style.display = 'block';
            };
        } catch (error) {
            console.error('HTML预览错误:', error);
            alert('HTML预览加载失败: ' + error.message);
        }
    }
}); 