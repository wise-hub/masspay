document.addEventListener('DOMContentLoaded', function() {
    const initialLang = localStorage.getItem('currentLang') || 'bg';
    loadLanguage(initialLang);
    document.getElementById('languageSelect').value = initialLang;
    updateFlagIcon(initialLang);
});

function changeLanguage() {
    const languageSelect = document.getElementById('languageSelect');
    const lang = languageSelect.value;
    updateFlagIcon(lang);
    loadLanguage(lang);
    localStorage.setItem('currentLang', lang); 
}

function updateFlagIcon(lang) {
    const flagIcon = document.getElementById('flagIcon');
    switch(lang) {
        case 'en':
            flagIcon.src = './masspay/static/uk.png';
            break;
        case 'bg':
            flagIcon.src = './masspay/static/bg.png';
            break;
    }
}

function loadLanguage(lang) {
    const cachedLangData = localStorage.getItem(`langData-${lang}`);
    if (cachedLangData) {
        console.log('Using cached language data');
        updateTexts(JSON.parse(cachedLangData));
    } else {
        console.log('Fetching language data from server');
        fetch(`static/lang.json`)
            .then(response => response.json())
            .then(data => {
                localStorage.setItem(`langData-${lang}`, JSON.stringify(data[lang]));
                updateTexts(data[lang]);
            })
            .catch(error => console.error('Failed to load language file:', error));
    }
}

function updateTexts(langData) {
    document.title = langData.title;
    document.getElementById('uploadHeading').textContent = langData.uploadHeading;
    document.getElementById('executionDate').previousElementSibling.textContent = langData.executionDateLabel;
    document.getElementById('iban').previousElementSibling.textContent = langData.ibanLabel;
    document.getElementById('companyName').previousElementSibling.textContent = langData.companyNameLabel;
    document.getElementById('fileInput').previousElementSibling.textContent = langData.selectFileLabel;
    document.querySelector('button').textContent = langData.uploadButton;
    window.downloadSuccess = langData.downloadSuccess;
    window.processSuccess = langData.processSuccess + ': ';
}

function uploadFile() {
    const responseMessage = document.getElementById('response-message');
    responseMessage.textContent = '';
    responseMessage.className = 'response-message';

    const formData = new FormData();
    formData.append('executionDate', document.getElementById('executionDate').value.replaceAll('-', ''));
    formData.append('iban', document.getElementById('iban').value);
    formData.append('companyName', document.getElementById('companyName').value.toUpperCase());
    formData.append('file', document.getElementById('fileInput').files[0]);

    fetch('./masspay/api/upload', {
        method: 'POST',
        body: formData,
    })
    .then(response => {
        const contentType = response.headers.get("content-type");
        if (contentType && contentType.includes("application/json")) {
            return response.json().then(data => {
                if (!data.success) {
                    throw new Error(data.msg);
                }
                return data;
            });
        } else if (contentType && contentType.includes("application/octet-stream")) {
            return response.blob();
        } else {
            throw new Error('Unsupported content type: ' + contentType);
        }
    })
    .then(data => {
        if (data instanceof Blob) {
            const url = window.URL.createObjectURL(data);
            const a = document.createElement('a');
            a.href = url;
            a.download = document.getElementById('fileInput').files[0].name; 
            document.body.appendChild(a);
            a.click();
            a.remove();
            window.URL.revokeObjectURL(url);
            responseMessage.textContent = window.downloadSuccess;
            responseMessage.classList.add('success');
        } else {
            responseMessage.textContent = window.processSuccess + data.message;
            responseMessage.classList.add('success');
        }
        
        setTimeout(() => {
            responseMessage.classList.remove('success');
            responseMessage.classList.remove('error');
            responseMessage.textContent = '';
        }, 10000); 
    })
    .catch(error => {
        responseMessage.textContent = 'Error uploading file: ' + error.message;
        responseMessage.classList.add('error');

        setTimeout(() => {
            responseMessage.classList.remove('success');
            responseMessage.classList.remove('error');
            responseMessage.textContent = '';
        }, 10000); 
    });
}
