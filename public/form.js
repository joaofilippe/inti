// ==============================
// Importação em Lote
// ==============================
var loteAtivo = -1;
var loteMandados = [];


const loteUpload = document.getElementById("lote-upload");
const loteStatus = document.getElementById("lote-status");
const loteFila = document.getElementById("lote-fila");
const loteLista = document.getElementById("lote-lista");
const loteTotal = document.getElementById("lote-total");
const btnLoteLimpar = document.getElementById("btn-lote-limpar");

const form = document.getElementById("mandadoForm");
const step1 = document.getElementById("step-1");

function preencherFormComLote(m) {
  if (m.Mandado) document.getElementById("mandado-id").value = m.Mandado;
  if (m.Nome) document.getElementById("nome-alvo").value = m.Nome;
  if (m.Documento) {
    const docEl = document.getElementById("documento");
    docEl.value = m.Documento;
    docEl.dispatchEvent(new Event("blur"));
  }
  if (m.Endereco) document.getElementById("endereco").value = m.Endereco;
  if (m.Cidade) document.getElementById("cidade").value = m.Cidade;
  if (m.Posicao) document.getElementById("posicao").value = m.Posicao;
  if (m.Sexo) {
    const rM = document.querySelector('input[name="sexo"][value="M"]');
    const rF = document.querySelector('input[name="sexo"][value="F"]');
    if (m.Sexo === "M" && rM) rM.checked = true;
    else if (m.Sexo === "F" && rF) rF.checked = true;
  }
  step1.classList.add("active");
  document.getElementById("mandadoForm").scrollIntoView({ behavior: "smooth" });
}

const btnProximo = document.getElementById("btn-proximo");

function proximoLoteIndex() {
  const itens = loteLista.querySelectorAll(".lote-item");
  for (let i = loteAtivo + 1; i < itens.length; i++) {
    if (!itens[i].classList.contains("concluido")) return i;
  }
  return -1;
}

function atualizarBotaoProximo() {
  const proximo = proximoLoteIndex();
  if (proximo === -1) {
    btnProximo.style.display = "none";
  } else {
    btnProximo.style.display = "inline-block";
    btnProximo.innerText = `Próximo → (${proximo + 1}/${loteMandados.length})`;
  }
}

btnProximo.addEventListener("click", () => {
  const proximo = proximoLoteIndex();
  if (proximo === -1) return;
  const itens = loteLista.querySelectorAll(".lote-item");
  itens.forEach((b) => b.classList.remove("ativo"));
  itens[proximo].classList.add("ativo");
  loteAtivo = proximo;
  preencherFormComLote(loteMandados[proximo]);
  atualizarBotaoProximo();
});

function renderLoteFila(mandados) {
  loteMandados = mandados;
  loteLista.innerHTML = "";
  mandados.forEach(function (m, i) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "lote-item";
    btn.dataset.index = i;

    const numSpan = document.createElement("span");
    numSpan.className = "lote-item-num";
    numSpan.textContent = i + 1;

    const idSpan = document.createElement("span");
    idSpan.className = "lote-item-id";
    idSpan.textContent = m.Mandado || "—";

    const nomeSpan = document.createElement("span");
    nomeSpan.className = "lote-item-nome";
    nomeSpan.textContent = m.Nome || "—";

    btn.appendChild(numSpan);
    btn.appendChild(idSpan);
    btn.appendChild(nomeSpan);

    btn.addEventListener("click", function () {
      loteLista.querySelectorAll(".lote-item").forEach(function (b) {
        b.classList.remove("ativo");
      });
      btn.classList.add("ativo");
      loteAtivo = i;
      preencherFormComLote(m);
    });

    loteLista.appendChild(btn);
  });
  loteTotal.textContent = mandados.length;
  loteFila.classList.add("active");
}

if (loteUpload) {
  loteUpload.addEventListener("change", function (e) {
    const file = e.target.files[0];
    if (!file) return;

    loteStatus.innerHTML =
      '<span style="color:#60a5fa;">⏳ Extraindo documentos...</span>';

    const formData = new FormData();
    formData.append("file", file);

    fetch("/api/batch/extract", { method: "POST", body: formData })
      .then(function (res) {
        if (!res.ok) throw new Error("Erro " + res.status);
        return res.json();
      })
      .then(function (data) {
        loteStatus.innerHTML =
          '<span style="color:var(--success-color);">✅ ' +
          data.length +
          " documento(s) lido(s).</span>";
        renderLoteFila(data);
        loteUpload.value = "";
      })
      .catch(function (err) {
        console.error(err);
        loteStatus.innerHTML =
          '<span style="color:var(--error-color);">❌ Falha na extração.</span>';
        loteUpload.value = "";
      });
  });
}

if (btnLoteLimpar) {
  btnLoteLimpar.addEventListener("click", function () {
    loteAtivo = -1;
    loteMandados = [];
    loteLista.innerHTML = "";
    loteFila.classList.remove("active");
    loteStatus.innerHTML = "";
    loteTotal.textContent = "0";
    btnProximo.style.display = "none";
  });
}

// ==============================
// Upload e Extração via IA
// ==============================
const uploadInput = document.getElementById("mandado-upload");
const uploadStatus = document.getElementById("upload-status");

if (uploadInput) {
  uploadInput.addEventListener("change", function (e) {
    const file = e.target.files[0];
    if (!file) return;

    uploadStatus.innerHTML =
      '<span style="color: #60a5fa;">⏳ Analisando documento com IA...</span>';

    const formData = new FormData();
    formData.append("file", file);

    fetch("/api/extract", { method: "POST", body: formData })
      .then((res) => {
        if (!res.ok)
          throw new Error("Falha na extração. Código: " + res.status);
        return res.json();
      })
      .then((data) => {
        uploadStatus.innerHTML =
          '<span style="color: var(--success-color);">✅ Documento lido com sucesso!</span>';

        if (data.Mandado)
          document.getElementById("mandado-id").value = data.Mandado;
        if (data.Nome) document.getElementById("nome-alvo").value = data.Nome;
        if (data.Documento) {
          const docEl = document.getElementById("documento");
          docEl.value = data.Documento;
          docEl.dispatchEvent(new Event("blur"));
        }
        if (data.Endereco)
          document.getElementById("endereco").value = data.Endereco;
        if (data.Cidade) document.getElementById("cidade").value = data.Cidade;
        if (data.Posicao)
          document.getElementById("posicao").value = data.Posicao;
        if (data.Sexo) {
          const rM = document.querySelector('input[name="sexo"][value="M"]');
          const rF = document.querySelector('input[name="sexo"][value="F"]');
          if ((data.Sexo === "M" || data.Sexo === "Masculino") && rM)
            rM.checked = true;
          else if ((data.Sexo === "F" || data.Sexo === "Feminino") && rF)
            rF.checked = true;
        }

        uploadInput.value = "";
      })
      .catch((err) => {
        console.error(err);
        uploadStatus.innerHTML =
          '<span style="color: var(--error-color);">❌ Erro na leitura. Verifique log.</span>';
        uploadInput.value = "";
      });
  });
}

// ==============================
// PJ — Representante e Máscara
// ==============================
const isPjCheck = document.getElementById("is-pj");
const blocoRep = document.getElementById("bloco-representante");
const docInput = document.getElementById("documento");
const lblDocumento = document.getElementById("lbl-documento");

isPjCheck.addEventListener("change", function () {
  blocoRep.style.display = this.checked ? "grid" : "none";
  docInput.placeholder = this.checked
    ? "Ex: 00.000.000/0000-00 (CNPJ)"
    : "Apenas números...";
  lblDocumento.innerText = this.checked
    ? "Documento (CNPJ)"
    : "Documento (CPF)";
  docInput.dispatchEvent(new Event("blur"));
});

// ==============================
// Máscaras
// ==============================
const mandadoInput = document.getElementById("mandado-id");
mandadoInput.addEventListener("blur", function () {
  let val = this.value.replace(/\D/g, "").substring(0, 7);
  if (val.length > 1) val = val.slice(0, -1) + "-" + val.slice(-1);
  this.value = val;
});

docInput.addEventListener("blur", function () {
  let val = this.value.replace(/\D/g, "");
  if (isPjCheck.checked) {
    if (val.length > 14) val = val.substring(0, 14);
    val = val.replace(/^(\d{2})(\d)/, "$1.$2");
    val = val.replace(/^(\d{2})\.(\d{3})(\d)/, "$1.$2.$3");
    val = val.replace(/\.(\d{3})(\d)/, ".$1/$2");
    val = val.replace(/(\d{4})(\d)/, "$1-$2");
  } else {
    if (val.length > 11) val = val.substring(0, 11);
    val = val.replace(/(\d{3})(\d)/, "$1.$2");
    val = val.replace(/(\d{3})(\d)/, "$1.$2");
    val = val.replace(/(\d{3})(\d{1,2})$/, "$1-$2");
  }
  this.value = val;
});

const repDocInput = document.getElementById("rep-doc");
repDocInput.addEventListener("input", function () {
  let val = this.value.replace(/\D/g, "").substring(0, 11);
  val = val.replace(/(\d{3})(\d)/, "$1.$2");
  val = val.replace(/(\d{3})(\d)/, "$1.$2");
  val = val.replace(/(\d{3})(\d{1,2})$/, "$1-$2");
  this.value = val;
});

function aplicarMascaraData(input) {
  input.addEventListener("input", function () {
    let val = this.value.replace(/\D/g, "").substring(0, 8);
    if (val.length > 4) val = val.replace(/(\d{2})(\d{2})(\d+)/, "$1/$2/$3");
    else if (val.length > 2) val = val.replace(/(\d{2})(\d+)/, "$1/$2");
    this.value = val;
  });

  input.addEventListener("blur", function () {
    let val = this.value;
    if (!val) return;
    if (val.endsWith("/")) val = val.slice(0, -1);
    const parts = val.split("/");
    if (parts.length === 2 && parts[0].length === 2 && parts[1].length === 2) {
      this.value = val + "/" + new Date().getFullYear();
    } else if (parts.length === 3 && parts[2].length === 2) {
      parts[2] = "20" + parts[2];
      this.value = parts.join("/");
    }
  });
}

aplicarMascaraData(document.getElementById("data-carga"));

function aplicarMascaraHora(input) {
  input.addEventListener("input", function () {
    let val = this.value.replace(/\D/g, "");
    if (val.length > 4) val = val.substring(0, 4);
    if (val.length >= 2 && parseInt(val.substring(0, 2)) > 23)
      val = "23" + val.substring(2);
    if (val.length >= 4 && parseInt(val.substring(2, 4)) > 59)
      val = val.substring(0, 2) + "59";
    if (val.length > 2) val = val.substring(0, 2) + ":" + val.substring(2, 4);
    this.value = val;
  });
}

// ==============================
// Diligências
// ==============================
const dilContainer = document.getElementById("diligencias-container");

function setupDilRow(row) {
  aplicarMascaraData(row.querySelector(".dil-data"));
  aplicarMascaraHora(row.querySelector(".dil-hora"));
  const btnRemove = row.querySelector(".btn-remove-dil");
  if (btnRemove) {
    btnRemove.addEventListener("click", () => {
      row.remove();
      atualizarBotoesRemover(".diligencia-row", ".btn-remove-dil");
    });
  }
}

setupDilRow(dilContainer.querySelector(".diligencia-row"));

document.getElementById("btn-add-dil").addEventListener("click", () => {
  const template = dilContainer.querySelector(".diligencia-row");
  const newRow = template.cloneNode(true);
  newRow.querySelector(".dil-data").value = "";
  newRow.querySelector(".dil-hora").value = "";
  dilContainer.appendChild(newRow);
  setupDilRow(newRow);
  atualizarBotoesRemover(".diligencia-row", ".btn-remove-dil");
  newRow.querySelector(".dil-data").focus();
});

// ==============================
// Atos
// ==============================
const atoContainer = document.getElementById("atos-container");

function setupAtoRow(row) {
  aplicarMascaraData(row.querySelector(".ato-data"));
  aplicarMascaraHora(row.querySelector(".ato-hora"));
  const btnRemove = row.querySelector(".btn-remove-ato");
  if (btnRemove) {
    btnRemove.addEventListener("click", () => {
      row.remove();
      atualizarBotoesRemover(".ato-row", ".btn-remove-ato");
    });
  }
}

setupAtoRow(atoContainer.querySelector(".ato-row"));

document.getElementById("btn-add-ato").addEventListener("click", () => {
  const template = atoContainer.querySelector(".ato-row");
  const newRow = template.cloneNode(true);
  newRow.querySelector(".ato-nome").value = "";
  newRow.querySelector(".ato-data").value = "";
  newRow.querySelector(".ato-hora").value = "";
  atoContainer.appendChild(newRow);
  setupAtoRow(newRow);
  atualizarBotoesRemover(".ato-row", ".btn-remove-ato");
  newRow.querySelector(".ato-nome").focus();
});

// ==============================
// Contatos
// ==============================
const contatosContainer = document.getElementById("contatos-container");

function setupContatoRow(row) {
  const select = row.querySelector(".tipo-contato-select");
  const input = row.querySelector(".contato-valor");
  const btnRemove = row.querySelector(".btn-remove-contato");

  select.addEventListener("change", (e) => {
    const isEmail = e.target.value === "EMAIL";
    input.type = isEmail ? "email" : "tel";
    input.placeholder = isEmail
      ? "Ex: email@dominio.com"
      : "Ex: (11) 99999-9999";
    if (isEmail) input.value = input.value.replace(/[^a-zA-Z0-9@._-]/g, "");
    else input.dispatchEvent(new Event("blur"));
  });

  input.addEventListener("blur", function () {
    if (select.value === "EMAIL") return;
    let val = this.value.replace(/\D/g, "");
    if (val.length > 11) val = val.substring(0, 11);
    if (val.length > 10)
      val = val.replace(/(\d{2})(\d{5})(\d{4})/, "($1) $2-$3");
    else if (val.length > 6)
      val = val.replace(/(\d{2})(\d{4})(\d{0,4})/, "($1) $2-$3");
    else if (val.length > 2) val = val.replace(/(\d{2})(\d{0,5})/, "($1) $2");
    this.value = val;
  });

  if (btnRemove) {
    btnRemove.addEventListener("click", () => {
      row.remove();
      atualizarBotoesRemover(".contato-row", ".btn-remove-contato");
    });
  }

  const checkTerceiro = row.querySelector(".is-terceiro-check");
  const blocoTerceiro = row.querySelector(".bloco-terceiro");
  if (checkTerceiro && blocoTerceiro) {
    checkTerceiro.addEventListener("change", function () {
      blocoTerceiro.style.display = this.checked ? "grid" : "none";
    });
  }
}

setupContatoRow(contatosContainer.querySelector(".contato-row"));

document.getElementById("btn-add-contato").addEventListener("click", () => {
  const template = contatosContainer.querySelector(".contato-row");
  const newRow = template.cloneNode(true);
  newRow.querySelector(".contato-valor").value = "";
  const checkTerceiro = newRow.querySelector(".is-terceiro-check");
  if (checkTerceiro) checkTerceiro.checked = false;
  const blocoTerceiro = newRow.querySelector(".bloco-terceiro");
  if (blocoTerceiro) blocoTerceiro.style.display = "none";
  const tNome = newRow.querySelector(".terceiro-nome");
  if (tNome) tNome.value = "";
  const tRel = newRow.querySelector(".terceiro-relacao");
  if (tRel) tRel.value = "";
  contatosContainer.appendChild(newRow);
  setupContatoRow(newRow);
  atualizarBotoesRemover(".contato-row", ".btn-remove-contato");
  newRow.querySelector(".contato-valor").focus();
});

function atualizarBotoesRemover(rowSel, btnSel) {
  const rows = document.querySelectorAll(rowSel);
  rows.forEach((row) => {
    const btn = row.querySelector(btnSel);
    if (btn) btn.style.display = rows.length > 1 ? "block" : "none";
  });
}

// ==============================
// Situação e Envio
// ==============================
const inputSituacao = document.getElementById("situacao");
const lblSituacao = document.getElementById("lbl-situacao");
const btnSave = document.getElementById("btn-save");

inputSituacao.addEventListener("input", (e) => {
  const val = e.target.value;
  if (val === "") {
    lblSituacao.innerHTML = "Situação";
    btnSave.innerText = "Aguardando Situação...";
    btnSave.style.backgroundColor = "var(--primary-color)";
  }

  if (val === "4") {
    lblSituacao.innerHTML =
      'Situação <span style="color: var(--success-color)">(Positivo)</span>';
  }

  btnSave.innerText = "Salvar";
  btnSave.style.backgroundColor = "var(--success-color)";
});

btnSave.addEventListener("click", () => {
  if (!form.checkValidity()) {
    form.reportValidity();
    return;
  }
  const sit = inputSituacao.value;
  if (sit !== "4" && sit !== "6") {
    alert(
      "Por favor, digite 4 para Positivo ou 6 para Negativo no campo de Situação.",
    );
    inputSituacao.focus();
    return;
  }
  form.dispatchEvent(new Event("submit"));
});

form.addEventListener("submit", function (e) {
  console.log("Submetendo...")
  e.preventDefault();

  const stringDilArr = [];
  document.querySelectorAll(".diligencia-row").forEach((row) => {
    let dDate = row.querySelector(".dil-data").value;
    const dHora = row.querySelector(".dil-hora").value;
    if (dDate.trim() !== "") {
      if (dDate.length >= 5) dDate = dDate.substring(0, 5);
      let combined = dDate;
      if (dHora.trim() !== "") combined += ` ${dHora.replace(":", "h")}`;
      stringDilArr.push(combined);
    }
  });

  const arrContatos = [];
  document.querySelectorAll(".contato-row").forEach((row) => {
    const tipoValue = row.querySelector(".tipo-contato-select").value;
    const valorInput = row.querySelector(".contato-valor").value;
    const isTerceiro = row.querySelector(".is-terceiro-check").checked;
    if (valorInput.trim() !== "") {
      arrContatos.push({
        Tipo: tipoValue,
        Valor: valorInput,
        IsTerceiro: isTerceiro,
        NomeTerceiro: isTerceiro
          ? row.querySelector(".terceiro-nome").value
          : "",
        RelacaoTerceiro: isTerceiro
          ? row.querySelector(".terceiro-relacao").value
          : "",
      });
    }
  });

  const arrAtos = [];
  document.querySelectorAll(".ato-row").forEach((row) => {
    const aNome = row.querySelector(".ato-nome").value;
    const aData = row.querySelector(".ato-data").value;
    const aHora = row.querySelector(".ato-hora").value;
    if (aNome.trim() !== "" || aData.trim() !== "") {
      arrAtos.push({ NomeAto: aNome, DataCumprimento: aData, Horario: aHora });
    }
  });

  const isPJ = document.getElementById("is-pj").checked;
  const payload = {
    Mandado: document.getElementById("mandado-id").value,
    Situacao: inputSituacao.value,
    DataCarga: document.getElementById("data-carga").value,
    Atos: arrAtos,
    Diligencias: stringDilArr.join(", "),
    IsPJ: isPJ,
    Contato: arrContatos,
    Cidade: document.getElementById("cidade").value,
    Endereco: document.getElementById("endereco").value,
    RepresentanteNome: isPJ ? document.getElementById("rep-nome").value : "",
    RepresentanteDoc: isPJ ? document.getElementById("rep-doc").value : "",
    Nome: document.getElementById("nome-alvo").value,
    Sexo: document.querySelector('input[name="sexo"]:checked').value,
    Posicao: document.getElementById("posicao").value,
    Documento: document.getElementById("documento").value,
    Obs: document.getElementById("obs-mandado").value,
  };

  const btn = document.getElementById("btn-save");
  btn.innerText = "Gerando Documento...";
  btn.style.opacity = "0.7";

  fetch("/api/mandados", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  })
    .then((res) => {
      if (!res.ok)
        throw new Error(
          "Erro na solicitação. O servidor retornou " + res.status,
        );
      return res.blob();
    })
    .then((blob) => {
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${payload.Nome}-${payload.Mandado}.docx`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);

      btn.innerText = "Concluir e Salvar no Sistema";
      btn.style.opacity = "1";

      if (typeof loteAtivo !== "undefined" && loteAtivo >= 0) {
        const itemAtivo = document.querySelector(
          "#lote-lista .lote-item.ativo",
        );
        if (itemAtivo) {
          itemAtivo.classList.remove("ativo");
          itemAtivo.classList.add("concluido");
          itemAtivo.querySelector(".lote-item-num").textContent =
            "✓ " +
            itemAtivo
              .querySelector(".lote-item-num")
              .textContent.replace("✓ ", "");
        }
        loteAtivo = -1;
      }
      atualizarBotaoProximo();
    })
    .catch((e) => {
      console.error(e);
      alert("Ocorreu um erro ao gerar o arquivo. Consulte o log.");
      btn.innerText = "Concluir e Salvar no Sistema";
      btn.style.opacity = "1";
    });
});
