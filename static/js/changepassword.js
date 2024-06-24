document.addEventListener('DOMContentLoaded', function() {
  document.getElementById("floatingInput").focus();

    // Prevent form submission when not all fields are validated
    (() => {
      'use strict'

      // Fetch all the forms we want to apply custom Bootstrap validation styles to
      const forms = document.querySelectorAll('.needs-validation')

      // Loop over them and prevent submission
      Array.from(forms).forEach(form => {
        form.addEventListener('submit', event => {
          event.preventDefault();
          if (!samePassword()) {
            const errorMessage = document.getElementById('error-message2');
            errorMessage.classList.remove('d-none');
          }
          else if (!form.checkValidity()) {
            event.stopPropagation()
          }
          else {
              getSalt().then(salt => {
              var encPass = encryptPassword(salt)
              fetchForm(encPass)
              })
          }

          form.classList.add('was-validated')
        }, false)
      })
    })()
});

function encryptPassword(salt) {
  var password = document.getElementById("floatingPassword").value;

  var encryptedPassword = CryptoJS.SHA256(password + salt).toString(CryptoJS.enc.Hex);
  return encryptedPassword
}

function samePassword() {
  var pass2 = document.getElementById("floatingPassword2");

  if (pass2 === null) {
    return true
  }
  var pass1 = document.getElementById("floatingPassword").value;
  pass2 = pass2.value;

  if (pass1 === pass2) {
    return true
  }
  return false
}

function fetchForm(password) {
  const username = document.getElementById("floatingInput").value;
  const form = document.getElementById("loginForm");

  fetch(form.action, {
    method: "PUT",
    redirect: "error",
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ username, password })
  })
    .then(response => {
      console.log(response.status);
      if (response.status === 403) {
        const errorMessage = document.getElementById('error-message');
        errorMessage.classList.remove('d-none');
      } else if (response.ok) {
        window.location.href = response.headers.get("Location");
      } else {
        // Handle other potential errors
        console.error('Login failed with status:', response.status);
        return response.json().then(data => {
          console.log(data.redirectUrl)
           window.location.href = data.redirectUrl;
        });
      }
    })
    .catch(error => {
      console.error('Error during fetch:', error);
    });
};

async function getSalt() {
  const username = document.getElementById("floatingInput").value;

  const response = await fetch("/api/salt", {
    method: "GET",
    headers: { "username": username },
  });

  if (!response.ok) {
    console.error('Getting salt failed with status: ', response.status)
  }
  else {
    var salt = await response.text();
    return salt
  }
}


